package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mychow/ai-git/internal/note"
	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage knowledge notes",
	Long: `Manage knowledge notes for model memory.

Examples:
  ai-git note create "Go 并发模式" --type knowledge --tags "go,concurrency"
  ai-git note list --type knowledge
  ai-git note search "并发"
  ai-git note link note-001 --code "worker.go:Process"`,
}

var noteCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new note",
	Args:  cobra.ExactArgs(1),
	RunE:  runNoteCreate,
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes",
	RunE:  runNoteList,
}

var noteGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get note details",
	Args:  cobra.ExactArgs(1),
	RunE:  runNoteGet,
}

var noteSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search notes",
	Args:  cobra.ExactArgs(1),
	RunE:  runNoteSearch,
}

var noteUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update note",
	Args:  cobra.ExactArgs(1),
	RunE:  runNoteUpdate,
}

var noteDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete note",
	Args:  cobra.ExactArgs(1),
	RunE:  runNoteDelete,
}

var noteLinkCmd = &cobra.Command{
	Use:   "link <id>",
	Short: "Link note to code or task",
	Args:  cobra.ExactArgs(1),
	RunE:  runNoteLink,
}

var noteExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all notes",
	RunE:  runNoteExport,
}

var (
	noteType     string
	noteTags     string
	noteContent  string
	noteCodeRef  string
	noteTaskRef  string
	noteFormat   string
)

func init() {
	rootCmd.AddCommand(noteCmd)

	noteCmd.AddCommand(noteCreateCmd)
	noteCmd.AddCommand(noteListCmd)
	noteCmd.AddCommand(noteGetCmd)
	noteCmd.AddCommand(noteSearchCmd)
	noteCmd.AddCommand(noteUpdateCmd)
	noteCmd.AddCommand(noteDeleteCmd)
	noteCmd.AddCommand(noteLinkCmd)
	noteCmd.AddCommand(noteExportCmd)

	noteCreateCmd.Flags().StringVar(&noteType, "type", "knowledge", "Note type (knowledge, context, log)")
	noteCreateCmd.Flags().StringVar(&noteTags, "tags", "", "Comma-separated tags")
	noteCreateCmd.Flags().StringVar(&noteContent, "content", "", "Note content")

	noteListCmd.Flags().StringVar(&noteType, "type", "", "Filter by type")

	noteUpdateCmd.Flags().StringVar(&noteContent, "content", "", "Update content")
	noteUpdateCmd.Flags().StringVar(&noteTags, "tags", "", "Update tags")

	noteLinkCmd.Flags().StringVar(&noteCodeRef, "code", "", "Code reference (file:symbol)")
	noteLinkCmd.Flags().StringVar(&noteTaskRef, "task", "", "Task ID")

	noteExportCmd.Flags().StringVar(&noteFormat, "format", "markdown", "Export format (markdown, json)")
}

func getNoteManager() (*note.Manager, error) {
	home, _ := os.UserHomeDir()
	notesDir := filepath.Join(home, ".ai-git", "notes")

	mgr := note.NewManager(notesDir)
	if err := mgr.Init(); err != nil {
		return nil, err
	}

	return mgr, nil
}

func runNoteCreate(cmd *cobra.Command, args []string) error {
	mgr, err := getNoteManager()
	if err != nil {
		return err
	}

	title := args[0]

	var noteTypeVal note.NoteType
	switch noteType {
	case "context":
		noteTypeVal = note.NoteTypeContext
	case "log":
		noteTypeVal = note.NoteTypeLog
	default:
		noteTypeVal = note.NoteTypeKnowledge
	}

	var opts []note.NoteOption

	if noteContent != "" {
		opts = append(opts, note.WithContent(noteContent))
	}

	if noteTags != "" {
		tags := strings.Split(noteTags, ",")
		for i, t := range tags {
			tags[i] = strings.TrimSpace(t)
		}
		opts = append(opts, note.WithTags(tags))
	}

	n, err := mgr.Create(title, noteTypeVal, opts...)
	if err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":         n.ID,
		"title":      n.Title,
		"type":       n.Type,
		"tags":       n.Tags,
		"message":    "Note created successfully",
	})
}

func runNoteList(cmd *cobra.Command, args []string) error {
	mgr, err := getNoteManager()
	if err != nil {
		return err
	}

	var noteTypeVal note.NoteType
	if noteType != "" {
		noteTypeVal = note.NoteType(noteType)
	}

	notes, err := mgr.List(noteTypeVal)
	if err != nil {
		return err
	}

	result := make([]map[string]interface{}, 0, len(notes))
	for _, n := range notes {
		result = append(result, map[string]interface{}{
			"id":        n.ID,
			"title":     n.Title,
			"type":      n.Type,
			"tags":      n.Tags,
			"created_at": n.CreatedAt,
		})
	}

	return utils.OutputSuccess(map[string]interface{}{
		"count": len(result),
		"notes": result,
	})
}

func runNoteGet(cmd *cobra.Command, args []string) error {
	mgr, err := getNoteManager()
	if err != nil {
		return err
	}

	id := args[0]
	n, err := mgr.Get(id)
	if err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":         n.ID,
		"title":      n.Title,
		"type":       n.Type,
		"content":    n.Content,
		"tags":       n.Tags,
		"links":      n.Links,
		"code_refs":  n.CodeRefs,
		"task_refs":  n.TaskRefs,
		"created_at": n.CreatedAt,
		"updated_at": n.UpdatedAt,
	})
}

func runNoteSearch(cmd *cobra.Command, args []string) error {
	mgr, err := getNoteManager()
	if err != nil {
		return err
	}

	query := args[0]
	notes, err := mgr.Search(query)
	if err != nil {
		return err
	}

	result := make([]map[string]interface{}, 0, len(notes))
	for _, n := range notes {
		result = append(result, map[string]interface{}{
			"id":        n.ID,
			"title":     n.Title,
			"type":      n.Type,
			"tags":      n.Tags,
			"relevance": "matched",
		})
	}

	return utils.OutputSuccess(map[string]interface{}{
		"query":  query,
		"count":  len(result),
		"notes":  result,
	})
}

func runNoteUpdate(cmd *cobra.Command, args []string) error {
	mgr, err := getNoteManager()
	if err != nil {
		return err
	}

	id := args[0]

	var opts []note.NoteOption

	if noteContent != "" {
		opts = append(opts, note.WithContent(noteContent))
	}

	if noteTags != "" {
		tags := strings.Split(noteTags, ",")
		for i, t := range tags {
			tags[i] = strings.TrimSpace(t)
		}
		opts = append(opts, note.WithTags(tags))
	}

	n, err := mgr.Update(id, opts...)
	if err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":      n.ID,
		"title":   n.Title,
		"message": "Note updated successfully",
	})
}

func runNoteDelete(cmd *cobra.Command, args []string) error {
	mgr, err := getNoteManager()
	if err != nil {
		return err
	}

	id := args[0]

	if err := mgr.Delete(id); err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":      id,
		"message": "Note deleted successfully",
	})
}

func runNoteLink(cmd *cobra.Command, args []string) error {
	mgr, err := getNoteManager()
	if err != nil {
		return err
	}

	id := args[0]

	if noteCodeRef != "" {
		if err := mgr.LinkCode(id, noteCodeRef); err != nil {
			return err
		}
	}

	if noteTaskRef != "" {
		if err := mgr.LinkTask(id, noteTaskRef); err != nil {
			return err
		}
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":        id,
		"code_ref":  noteCodeRef,
		"task_ref":  noteTaskRef,
		"message":   "Note linked successfully",
	})
}

func runNoteExport(cmd *cobra.Command, args []string) error {
	mgr, err := getNoteManager()
	if err != nil {
		return err
	}

	content, err := mgr.Export(noteFormat)
	if err != nil {
		return err
	}

	fmt.Println(content)
	return nil
}
