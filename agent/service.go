package agent

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/copilot-extensions/rag-extension/copilot"
	"github.com/copilot-extensions/rag-extension/embedding"
	"github.com/copilot-extensions/rag-extension/prompt"
	"github.com/copilot-extensions/rag-extension/search"
)

// Service provides and endpoint for this agent to perform chat completions
type Service struct {
	pubKey *ecdsa.PublicKey

	// Singleton
	datasets     []*embedding.Dataset
	datasetsInit *sync.Once
}

func NewService(pubKey *ecdsa.PublicKey) *Service {
	return &Service{
		pubKey:       pubKey,
		datasetsInit: &sync.Once{},
	}
}

func (s *Service) ChatCompletion(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("Github-Public-Key-Signature")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to read request body: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Make sure the payload matches the signature. In this way, you can be sure
	// that an incoming request comes from github
	isValid, err := validPayload(body, sig, s.pubKey)
	if err != nil {
		fmt.Printf("failed to validate payload signature: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isValid {
		http.Error(w, "invalid payload signature", http.StatusUnauthorized)
		return
	}

	apiToken := r.Header.Get("X-GitHub-Token")
	integrationID := r.Header.Get("Copilot-Integration-Id")

	var req *copilot.ChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		fmt.Printf("failed to unmarshal request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.generateCompletion(r.Context(), integrationID, apiToken, req, w); err != nil {
		fmt.Printf("failed to execute agent: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Service) generateCompletion(ctx context.Context, integrationID, apiToken string, req *copilot.ChatRequest, w io.Writer) error {
	var messages []copilot.ChatMessage

	// Create embeddings from user messages
	for i := len(req.Messages) - 1; i >= 0; i-- {
		msg := req.Messages[i]
		if msg.Role != "user" {
			continue
		}

		// replace the tsp with typespec
		msg.Content = strings.ReplaceAll(msg.Content, " tsp", " typespec")

		// Filter empty messages
		if msg.Content == "" {
			continue
		}
		println("msg:", msg.Content)
		results, err := search.SearchTopKRelatedDocuments(msg.Content, 10)
		if err != nil {
			return fmt.Errorf("failed to search for related documents: %w", err)
		}
		files := make(map[string]bool)
		mergedChunks := make([]search.Index, 0)
		for _, result := range results {
			if files[result.Title] {
				continue
			}
			files[result.Title] = true
			mergedChunks = append(mergedChunks, search.Index{
				Title: result.Title,
			})
			if len(files) == 5 {
				break
			}
		}
		for i, _ := range mergedChunks {
			mergedChunks[i] = completeChunk(mergedChunks[i])
		}
		chunks := make([]string, 0)
		chunkLength := 0
		for _, result := range mergedChunks {
			chunk := fmt.Sprintf("- chunk_title: %s\n", result.Title)
			chunk += fmt.Sprintf("- chunk_link: %s\n", search.GetIndexLink(result))
			chunk += fmt.Sprintf("- content: %s\n", result.Chunk)
			chunkLength += len(chunk)
			if chunkLength > 100000 {
				break
			}
			chunks = append(chunks, chunk)
		}
		context := strings.Join(chunks, "-------------------------\n")
		prompt := prompt.BuildAnswerPrompt(msg.Content, context)
		println(fmt.Printf("message: %s\ncontext:\n%s", msg.Content, context))
		messages = append(messages, copilot.ChatMessage{
			Role:    "system",
			Content: prompt,
		})
		break
	}

	messages = append(messages, req.Messages...)

	chatReq := &copilot.ChatCompletionsRequest{
		Model:               copilot.ModelGPT4o,
		Messages:            messages,
		Stream:              true,
		Temperature:         0.1,
		TopP:                0.1,
		MaxCompletionTokens: 1024,
	}

	stream, err := copilot.ChatCompletions(ctx, "copilot-chat", apiToken, chatReq)
	if err != nil {
		return fmt.Errorf("failed to get chat completions stream: %w", err)
	}
	defer stream.Close()

	reader := bufio.NewScanner(stream)
	for reader.Scan() {
		buf := reader.Bytes()
		_, err := w.Write(buf)
		if err != nil {
			return fmt.Errorf("failed to write to stream: %w", err)
		}

		if _, err := w.Write([]byte("\n")); err != nil {
			return fmt.Errorf("failed to write delimiter to stream: %w", err)
		}
	}

	if err := reader.Err(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return fmt.Errorf("failed to read from stream: %w", err)
	}
	println("done")

	return nil
}

func completeChunk(chunk search.Index) search.Index {
	chunks, err := search.GetCompleteContext(chunk)
	if err != nil {
		return chunk
	}
	if len(chunks) == 0 {
		return chunk
	}
	var contents []string
	totalLength := 0
	for _, chunk := range chunks {
		if totalLength+len(chunk.Chunk) > 10000 {
			break
		}
		totalLength += len(chunk.Chunk)
		contents = append(contents, chunk.Chunk)
	}
	chunk.Chunk = strings.Join(contents, "\n")
	chunk.Title = chunks[0].Title
	chunk.Header1 = chunks[0].Header1
	chunk.Header2 = ""
	chunk.Header3 = ""
	chunk.OrdinalPosition = 0
	return chunk
}

func GetAllFilesInDir(root string) ([]string, error) {
	var result []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

// asn1Signature is a struct for ASN.1 serializing/parsing signatures.
type asn1Signature struct {
	R *big.Int
	S *big.Int
}

func validPayload(data []byte, sig string, publicKey *ecdsa.PublicKey) (bool, error) {
	asnSig, err := base64.StdEncoding.DecodeString(sig)
	parsedSig := asn1Signature{}
	if err != nil {
		return false, err
	}
	rest, err := asn1.Unmarshal(asnSig, &parsedSig)
	if err != nil || len(rest) != 0 {
		return false, err
	}

	// Verify the SHA256 encoded payload against the signature with GitHub's Key
	digest := sha256.Sum256(data)
	return ecdsa.Verify(publicKey, digest[:], parsedSig.R, parsedSig.S), nil
}

type AgentResponse struct {
	FindAnswer         bool     `json:"find_answer"`
	Answer             string   `json:"answer"`
	NeedFullContext    bool     `json:"need_full_context"`
	NeedFullContextIDs []string `json:"need_full_context_chunk_ids"`
	ReferenceChunkIDs  []string `json:"reference_chunk_ids"`
}
