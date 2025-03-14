package prompt

import (
	"os"
	"path/filepath"
	"strings"
)

// read prompt_template/answer_question.md, replace {{context}} with question and context
func BuildAnswerPrompt(message, context string) string {
	template := filepath.Join("prompt", "prompt_template", "answer_question.md")
	absPath, err := filepath.Abs(template)
	if err != nil {
		panic(err)
	}
	content, err := os.ReadFile(absPath)
	if err != nil {
		panic(err)
	}
	prompt := string(content)
	prompt = replacePlaceholder(prompt, "{{context}}", context)
	prompt = replacePlaceholder(prompt, "{{message}}", message)
	return prompt
}

func replacePlaceholder(template, placeholder, value string) string {
	return strings.ReplaceAll(template, placeholder, value)
}
