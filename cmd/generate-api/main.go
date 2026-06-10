package main

import (
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type apiName struct {
	Input  string
	Snake  string
	Kebab  string
	Pascal string
}

func main() {
	nameFlag := flag.String("name", "", "api name, for example: user-profile")
	methodFlag := flag.String("method", "GET", "http method: GET, POST, PUT, PATCH, DELETE")
	pathFlag := flag.String("path", "", "route path, for example: /api/user-profile")
	flag.Parse()

	name, err := parseAPIName(*nameFlag)
	if err != nil {
		exit(err)
	}

	method, err := parseMethod(*methodFlag)
	if err != nil {
		exit(err)
	}

	path := strings.TrimSpace(*pathFlag)
	if path == "" {
		path = "/api/" + name.Kebab
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if err := generate(name, method, path); err != nil {
		exit(err)
	}

	fmt.Printf("generated api %q\n", name.Kebab)
	fmt.Printf("- internal/http/handler/%s.go\n", name.Snake)
	fmt.Printf("- %s %s in internal/http/routes.go\n", method, path)
}

func generate(name apiName, method string, path string) error {
	handlerPath := filepath.Join("internal", "http", "handler", name.Snake+".go")
	if _, err := os.Stat(handlerPath); err == nil {
		return fmt.Errorf("handler file already exists: %s", handlerPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := writeGoFile(handlerPath, handlerFile(name)); err != nil {
		return err
	}

	routeEntry := fmt.Sprintf("\n\t%sHandler := handler.New%sHandler(deps, logger.With(\"component\", \"api\", \"api\", %q))\n\tapp.%s(%q, %sHandler.Handle)\n", name.Snake, name.Pascal, name.Kebab, titleMethod(method), path, name.Snake)
	if err := insertBefore("internal/http/routes.go", "\n\t// generator:api-route", routeEntry); err != nil {
		return err
	}

	return gofmtFiles("internal/http/routes.go", handlerPath)
}

func parseAPIName(input string) (apiName, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return apiName{}, fmt.Errorf("name is required")
	}

	parts := splitName(input)
	if len(parts) == 0 {
		return apiName{}, fmt.Errorf("name must contain letters or numbers")
	}

	for _, part := range parts {
		if !unicode.IsLetter([]rune(part)[0]) {
			return apiName{}, fmt.Errorf("each name segment must start with a letter: %q", part)
		}
	}

	return apiName{
		Input:  input,
		Snake:  strings.Join(parts, "_"),
		Kebab:  strings.Join(parts, "-"),
		Pascal: pascal(parts),
	}, nil
}

func parseMethod(input string) (string, error) {
	method := strings.ToUpper(strings.TrimSpace(input))
	if method == "" {
		method = "GET"
	}

	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
		return method, nil
	default:
		return "", fmt.Errorf("unsupported method %q", input)
	}
}

func titleMethod(method string) string {
	return strings.ToUpper(method[:1]) + strings.ToLower(method[1:])
}

func splitName(input string) []string {
	var parts []string
	var current []rune

	flush := func() {
		if len(current) == 0 {
			return
		}
		parts = append(parts, strings.ToLower(string(current)))
		current = nil
	}

	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current = append(current, r)
			continue
		}
		flush()
	}
	flush()

	return parts
}

func pascal(parts []string) string {
	var builder strings.Builder
	for _, part := range parts {
		runes := []rune(part)
		builder.WriteRune(unicode.ToUpper(runes[0]))
		builder.WriteString(string(runes[1:]))
	}

	return builder.String()
}

func handlerFile(name apiName) string {
	return fmt.Sprintf(`package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"template_sch/internal/http/dependency"
)

type %[1]sHandler struct {
	deps   dependency.Dependencies
	logger *slog.Logger
}

func New%[1]sHandler(deps dependency.Dependencies, logger *slog.Logger) %[1]sHandler {
	return %[1]sHandler{
		deps:   deps,
		logger: logger,
	}
}

func (h %[1]sHandler) Handle(c *fiber.Ctx) error {
	// Ganti isi method ini dengan logic utama API Anda.
	if h.deps.DB == nil || h.deps.DB.Primary == nil {
		h.logger.Debug("primary database is disabled")
	}

	h.logger.Info("%[2]s api handled")
	return c.JSON(fiber.Map{
		"status": "ok",
		"api":    "%[2]s",
	})
}
`, name.Pascal, name.Kebab)
}

func insertBefore(path string, marker string, insertion string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	text := string(content)
	if strings.Contains(text, insertion) {
		return nil
	}

	index := strings.Index(text, marker)
	if index < 0 {
		return fmt.Errorf("marker %q not found in %s", marker, path)
	}

	next := text[:index] + insertion + text[index:]
	return os.WriteFile(path, []byte(next), 0644)
}

func writeGoFile(path string, content string) error {
	formatted, err := format.Source([]byte(content))
	if err != nil {
		return err
	}

	return os.WriteFile(path, formatted, 0644)
}

func gofmtFiles(paths ...string) error {
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		formatted, err := format.Source(content)
		if err != nil {
			return fmt.Errorf("format %s: %w", path, err)
		}

		if err := os.WriteFile(path, formatted, 0644); err != nil {
			return err
		}
	}

	return nil
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, "generate-api:", err)
	os.Exit(1)
}
