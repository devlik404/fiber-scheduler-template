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

const defaultSchedule = "*/1 * * * *"

type jobName struct {
	Input  string
	Snake  string
	Kebab  string
	Pascal string
	Env    string
}

func main() {
	nameFlag := flag.String("name", "", "job name, for example: invoice-sync")
	scheduleFlag := flag.String("schedule", defaultSchedule, "cron schedule, for example: */5 * * * *")
	flag.Parse()

	name, err := parseJobName(*nameFlag)
	if err != nil {
		exit(err)
	}

	schedule := strings.TrimSpace(*scheduleFlag)
	if schedule == "" {
		schedule = defaultSchedule
	}

	if err := generate(name, schedule); err != nil {
		exit(err)
	}

	fmt.Printf("generated job %q\n", name.Kebab)
	fmt.Printf("- internal/task/jobs/%s.go\n", name.Snake)
	fmt.Printf("- %s_SCHEDULE in .env.example and config\n", name.Env)
}

func generate(name jobName, schedule string) error {
	jobPath := filepath.Join("internal", "task", "jobs", name.Snake+".go")
	if _, err := os.Stat(jobPath); err == nil {
		return fmt.Errorf("job file already exists: %s", jobPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := writeGoFile(jobPath, jobFile(name)); err != nil {
		return err
	}

	if err := insertBefore(".env.example", "# generator:task-schedule-env", fmt.Sprintf("%s_SCHEDULE=%s\n", name.Env, schedule)); err != nil {
		return err
	}

	if err := insertBefore("internal/config/config.go", "\t// generator:task-config-field", fmt.Sprintf("\t%sSchedule string\n", name.Pascal)); err != nil {
		return err
	}

	configLoad := fmt.Sprintf("\t\t\t%sSchedule: stringEnv(\"%s_SCHEDULE\", %q),\n", name.Pascal, name.Env, schedule)
	if err := insertBefore("internal/config/config.go", "\t\t\t// generator:task-config-load", configLoad); err != nil {
		return err
	}

	registryEntry := fmt.Sprintf("\t\t{\n\t\t\tName:     %q,\n\t\t\tSchedule: r.cfg.%sSchedule,\n\t\t\tHandler:  jobs.New%sJob(r.deps.DB, r.logger.With(\"component\", \"job\", \"job\", %q)),\n\t\t},\n", name.Kebab+"-task", name.Pascal, name.Pascal, name.Kebab+"-task")
	if err := insertBefore("internal/task/registry.go", "\t\t// generator:task-registry-job", registryEntry); err != nil {
		return err
	}

	return gofmtFiles("internal/config/config.go", "internal/task/registry.go", jobPath)
}

func parseJobName(input string) (jobName, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return jobName{}, fmt.Errorf("name is required")
	}

	parts := splitName(input)
	if len(parts) == 0 {
		return jobName{}, fmt.Errorf("name must contain letters or numbers")
	}

	for _, part := range parts {
		if !unicode.IsLetter([]rune(part)[0]) {
			return jobName{}, fmt.Errorf("each name segment must start with a letter: %q", part)
		}
	}

	return jobName{
		Input:  input,
		Snake:  strings.Join(parts, "_"),
		Kebab:  strings.Join(parts, "-"),
		Pascal: pascal(parts),
		Env:    strings.ToUpper(strings.Join(parts, "_")),
	}, nil
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

func jobFile(name jobName) string {
	return fmt.Sprintf(`package jobs

import (
	"context"
	"log/slog"

	"template_sch/internal/platform/database"
)

type %[1]sJob struct {
	db     *database.Connections
	logger *slog.Logger
}

func New%[1]sJob(db *database.Connections, logger *slog.Logger) %[1]sJob {
	return %[1]sJob{
		db:     db,
		logger: logger,
	}
}

func (j %[1]sJob) Run(ctx context.Context) error {
	// Ganti isi method ini dengan logic utama scheduler Anda.
	if j.db == nil || j.db.Primary == nil {
		j.logger.Debug("primary database is disabled")
	}

	j.logger.Info("%[2]s task logic executed")
	return nil
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
	fmt.Fprintln(os.Stderr, "generate-job:", err)
	os.Exit(1)
}
