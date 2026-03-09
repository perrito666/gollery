// Command gollery-users manages users.json and album.json files for gollery.
//
// User management:
//
//	gollery-users -file users.json list
//	gollery-users -file users.json add -username alice -password secret
//	gollery-users -file users.json add -username alice -password secret -admin -groups editors,viewers
//	gollery-users -file users.json remove -username alice
//	gollery-users -file users.json passwd -username alice -password newpass
//	gollery-users -file users.json set-admin -username alice -admin
//	gollery-users -file users.json set-admin -username alice -admin=false
//	gollery-users -file users.json set-groups -username alice -groups editors,viewers
//	gollery-users -file users.json add-groups -username alice -groups newgroup1,newgroup2
//	gollery-users -file users.json remove-groups -username alice -groups oldgroup
//
// Album management:
//
//	gollery-users init-album -dir /path/to/album -title "My Album"
//	gollery-users init-album -dir /path/to/album -title "Private" -access restricted -allowed-users alice,bob
//	gollery-users init-album -dir /path/to/album -title "Members Only" -access authenticated
//
// Validated editing (visudo-style):
//
//	gollery-users edit users -file users.json
//	gollery-users edit album -file /path/to/album.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// userEntry mirrors auth.UserEntry so this tool has no internal dependencies.
type userEntry struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Groups   []string `json:"groups"`
	IsAdmin  bool     `json:"is_admin"`
}

// albumConfig mirrors config.AlbumConfig for album.json creation/validation.
type albumConfig struct {
	Title       string        `json:"title,omitempty"`
	Description string        `json:"description,omitempty"`
	Inherit     *bool         `json:"inherit,omitempty"`
	Access      *accessConfig `json:"access,omitempty"`
}

type accessConfig struct {
	View          string   `json:"view,omitempty"`
	AllowedUsers  []string `json:"allowed_users,omitempty"`
	AllowedGroups []string `json:"allowed_groups,omitempty"`
	Admins        []string `json:"admins,omitempty"`
}

var validAccessModes = map[string]bool{
	"public":        true,
	"authenticated": true,
	"restricted":    true,
}

func main() {
	filePath := flag.String("file", "users.json", "path to users.json")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "list":
		cmdList(*filePath)
	case "add":
		cmdAdd(*filePath, cmdArgs)
	case "remove":
		cmdRemove(*filePath, cmdArgs)
	case "passwd":
		cmdPasswd(*filePath, cmdArgs)
	case "set-admin":
		cmdSetAdmin(*filePath, cmdArgs)
	case "set-groups":
		cmdSetGroups(*filePath, cmdArgs)
	case "add-groups":
		cmdAddGroups(*filePath, cmdArgs)
	case "remove-groups":
		cmdRemoveGroups(*filePath, cmdArgs)
	case "init-album":
		cmdInitAlbum(cmdArgs)
	case "edit":
		cmdEdit(*filePath, cmdArgs)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: gollery-users -file <users.json> <command> [flags]

User commands:
  list                                   List all users
  add       -username U -password P      Add a new user (optional: -admin, -groups g1,g2)
  remove    -username U                  Remove a user
  passwd    -username U -password P      Change a user's password
  set-admin -username U -admin           Set or clear admin flag (-admin or -admin=false)
  set-groups  -username U -groups G      Replace user groups (comma-separated, empty to clear)
  add-groups  -username U -groups G      Add groups to a user (comma-separated)
  remove-groups -username U -groups G    Remove groups from a user (comma-separated)

Album commands:
  init-album -dir D -title T             Create album.json (optional: -description, -access,
                                          -allowed-users, -allowed-groups, -admins, -no-inherit)

Editing (visudo-style, uses $EDITOR):
  edit users                             Edit and validate users.json
  edit album -file /path/to/album.json   Edit and validate album.json
`)
}

func loadUsers(path string) ([]userEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var users []userEntry
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return users, nil
}

func saveUsers(path string, users []userEntry) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0600)
}

func hashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func findUser(users []userEntry, username string) int {
	for i, u := range users {
		if u.Username == username {
			return i
		}
	}
	return -1
}

func cmdList(filePath string) {
	users, err := loadUsers(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(users) == 0 {
		fmt.Println("No users found.")
		return
	}
	for _, u := range users {
		admin := ""
		if u.IsAdmin {
			admin = " [admin]"
		}
		groups := ""
		if len(u.Groups) > 0 {
			groups = " groups=" + strings.Join(u.Groups, ",")
		}
		fmt.Printf("  %s%s%s\n", u.Username, admin, groups)
	}
}

func cmdAdd(filePath string, args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	username := fs.String("username", "", "username (required)")
	password := fs.String("password", "", "plaintext password (required)")
	admin := fs.Bool("admin", false, "grant admin privileges")
	groups := fs.String("groups", "", "comma-separated group list")
	fs.Parse(args)

	if *username == "" || *password == "" {
		fmt.Fprintf(os.Stderr, "error: -username and -password are required\n")
		os.Exit(1)
	}

	users, err := loadUsers(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if findUser(users, *username) >= 0 {
		fmt.Fprintf(os.Stderr, "error: user %q already exists\n", *username)
		os.Exit(1)
	}

	hash, err := hashPassword(*password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error hashing password: %v\n", err)
		os.Exit(1)
	}

	var groupList []string
	if *groups != "" {
		groupList = strings.Split(*groups, ",")
	}

	users = append(users, userEntry{
		Username: *username,
		Password: hash,
		Groups:   groupList,
		IsAdmin:  *admin,
	})

	if err := saveUsers(filePath, users); err != nil {
		fmt.Fprintf(os.Stderr, "error saving: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("User %q added.\n", *username)
}

func cmdRemove(filePath string, args []string) {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	username := fs.String("username", "", "username to remove (required)")
	fs.Parse(args)

	if *username == "" {
		fmt.Fprintf(os.Stderr, "error: -username is required\n")
		os.Exit(1)
	}

	users, err := loadUsers(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	idx := findUser(users, *username)
	if idx < 0 {
		fmt.Fprintf(os.Stderr, "error: user %q not found\n", *username)
		os.Exit(1)
	}

	users = append(users[:idx], users[idx+1:]...)

	if err := saveUsers(filePath, users); err != nil {
		fmt.Fprintf(os.Stderr, "error saving: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("User %q removed.\n", *username)
}

func cmdPasswd(filePath string, args []string) {
	fs := flag.NewFlagSet("passwd", flag.ExitOnError)
	username := fs.String("username", "", "username (required)")
	password := fs.String("password", "", "new plaintext password (required)")
	fs.Parse(args)

	if *username == "" || *password == "" {
		fmt.Fprintf(os.Stderr, "error: -username and -password are required\n")
		os.Exit(1)
	}

	users, err := loadUsers(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	idx := findUser(users, *username)
	if idx < 0 {
		fmt.Fprintf(os.Stderr, "error: user %q not found\n", *username)
		os.Exit(1)
	}

	hash, err := hashPassword(*password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error hashing password: %v\n", err)
		os.Exit(1)
	}

	users[idx].Password = hash

	if err := saveUsers(filePath, users); err != nil {
		fmt.Fprintf(os.Stderr, "error saving: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Password updated for %q.\n", *username)
}

func cmdSetAdmin(filePath string, args []string) {
	fs := flag.NewFlagSet("set-admin", flag.ExitOnError)
	username := fs.String("username", "", "username (required)")
	admin := fs.Bool("admin", false, "admin status")
	fs.Parse(args)

	if *username == "" {
		fmt.Fprintf(os.Stderr, "error: -username is required\n")
		os.Exit(1)
	}

	users, err := loadUsers(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	idx := findUser(users, *username)
	if idx < 0 {
		fmt.Fprintf(os.Stderr, "error: user %q not found\n", *username)
		os.Exit(1)
	}

	users[idx].IsAdmin = *admin

	if err := saveUsers(filePath, users); err != nil {
		fmt.Fprintf(os.Stderr, "error saving: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Admin status for %q set to %v.\n", *username, *admin)
}

func cmdSetGroups(filePath string, args []string) {
	fs := flag.NewFlagSet("set-groups", flag.ExitOnError)
	username := fs.String("username", "", "username (required)")
	groups := fs.String("groups", "", "comma-separated group list (empty to clear)")
	fs.Parse(args)

	if *username == "" {
		fmt.Fprintf(os.Stderr, "error: -username is required\n")
		os.Exit(1)
	}

	users, err := loadUsers(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	idx := findUser(users, *username)
	if idx < 0 {
		fmt.Fprintf(os.Stderr, "error: user %q not found\n", *username)
		os.Exit(1)
	}

	var groupList []string
	if *groups != "" {
		groupList = strings.Split(*groups, ",")
	}
	users[idx].Groups = groupList

	if err := saveUsers(filePath, users); err != nil {
		fmt.Fprintf(os.Stderr, "error saving: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Groups for %q set to %v.\n", *username, groupList)
}

func cmdAddGroups(filePath string, args []string) {
	fs := flag.NewFlagSet("add-groups", flag.ExitOnError)
	username := fs.String("username", "", "username (required)")
	groups := fs.String("groups", "", "comma-separated groups to add (required)")
	fs.Parse(args)

	if *username == "" || *groups == "" {
		fmt.Fprintf(os.Stderr, "error: -username and -groups are required\n")
		os.Exit(1)
	}

	users, err := loadUsers(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	idx := findUser(users, *username)
	if idx < 0 {
		fmt.Fprintf(os.Stderr, "error: user %q not found\n", *username)
		os.Exit(1)
	}

	existing := make(map[string]bool, len(users[idx].Groups))
	for _, g := range users[idx].Groups {
		existing[g] = true
	}

	toAdd := strings.Split(*groups, ",")
	var added []string
	for _, g := range toAdd {
		g = strings.TrimSpace(g)
		if g != "" && !existing[g] {
			users[idx].Groups = append(users[idx].Groups, g)
			existing[g] = true
			added = append(added, g)
		}
	}

	if len(added) == 0 {
		fmt.Printf("No new groups to add for %q (already a member).\n", *username)
		return
	}

	if err := saveUsers(filePath, users); err != nil {
		fmt.Fprintf(os.Stderr, "error saving: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Added groups %v to %q. Current groups: %v\n", added, *username, users[idx].Groups)
}

func cmdRemoveGroups(filePath string, args []string) {
	fs := flag.NewFlagSet("remove-groups", flag.ExitOnError)
	username := fs.String("username", "", "username (required)")
	groups := fs.String("groups", "", "comma-separated groups to remove (required)")
	fs.Parse(args)

	if *username == "" || *groups == "" {
		fmt.Fprintf(os.Stderr, "error: -username and -groups are required\n")
		os.Exit(1)
	}

	users, err := loadUsers(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	idx := findUser(users, *username)
	if idx < 0 {
		fmt.Fprintf(os.Stderr, "error: user %q not found\n", *username)
		os.Exit(1)
	}

	toRemove := make(map[string]bool)
	for _, g := range strings.Split(*groups, ",") {
		g = strings.TrimSpace(g)
		if g != "" {
			toRemove[g] = true
		}
	}

	var kept []string
	var removed []string
	for _, g := range users[idx].Groups {
		if toRemove[g] {
			removed = append(removed, g)
		} else {
			kept = append(kept, g)
		}
	}

	if len(removed) == 0 {
		fmt.Printf("No matching groups to remove for %q.\n", *username)
		return
	}

	users[idx].Groups = kept

	if err := saveUsers(filePath, users); err != nil {
		fmt.Fprintf(os.Stderr, "error saving: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Removed groups %v from %q. Current groups: %v\n", removed, *username, kept)
}

func cmdInitAlbum(args []string) {
	fs := flag.NewFlagSet("init-album", flag.ExitOnError)
	dir := fs.String("dir", "", "directory to create album.json in (required)")
	title := fs.String("title", "", "album title (required)")
	description := fs.String("description", "", "album description")
	accessMode := fs.String("access", "", "access mode: public, authenticated, or restricted")
	allowedUsers := fs.String("allowed-users", "", "comma-separated allowed users (for restricted)")
	allowedGroups := fs.String("allowed-groups", "", "comma-separated allowed groups (for restricted)")
	admins := fs.String("admins", "", "comma-separated album admins")
	noInherit := fs.Bool("no-inherit", false, "disable config inheritance from parent")
	fs.Parse(args)

	if *dir == "" || *title == "" {
		fmt.Fprintf(os.Stderr, "error: -dir and -title are required\n")
		os.Exit(1)
	}

	albumPath := *dir + "/album.json"

	// Check if album.json already exists.
	if _, err := os.Stat(albumPath); err == nil {
		fmt.Fprintf(os.Stderr, "error: %s already exists (use 'edit album -file %s' to modify)\n", albumPath, albumPath)
		os.Exit(1)
	}

	// Validate access mode if provided.
	if *accessMode != "" && !validAccessModes[*accessMode] {
		fmt.Fprintf(os.Stderr, "error: invalid access mode %q (must be public, authenticated, or restricted)\n", *accessMode)
		os.Exit(1)
	}

	cfg := albumConfig{
		Title:       *title,
		Description: *description,
	}

	if *noInherit {
		f := false
		cfg.Inherit = &f
	}

	if *accessMode != "" {
		cfg.Access = &accessConfig{
			View: *accessMode,
		}
		if *allowedUsers != "" {
			cfg.Access.AllowedUsers = splitCSV(*allowedUsers)
		}
		if *allowedGroups != "" {
			cfg.Access.AllowedGroups = splitCSV(*allowedGroups)
		}
		if *admins != "" {
			cfg.Access.Admins = splitCSV(*admins)
		}
	}

	// Ensure directory exists.
	if err := os.MkdirAll(*dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating directory: %v\n", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling: %v\n", err)
		os.Exit(1)
	}
	data = append(data, '\n')

	if err := os.WriteFile(albumPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s\n", albumPath)
}

func cmdEdit(defaultFilePath string, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: edit requires a type: 'users' or 'album'\n")
		usage()
		os.Exit(1)
	}

	editType := args[0]
	editArgs := args[1:]

	switch editType {
	case "users":
		cmdEditUsers(defaultFilePath, editArgs)
	case "album":
		cmdEditAlbum(editArgs)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown edit type %q (must be 'users' or 'album')\n", editType)
		os.Exit(1)
	}
}

func cmdEditUsers(defaultFilePath string, args []string) {
	fs := flag.NewFlagSet("edit-users", flag.ExitOnError)
	filePath := fs.String("file", "", "path to users.json (overrides -file)")
	fs.Parse(args)

	path := defaultFilePath
	if *filePath != "" {
		path = *filePath
	}

	editAndValidate(path, func(data []byte) error {
		var users []userEntry
		if err := json.Unmarshal(data, &users); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		// Validate structure.
		seen := make(map[string]bool)
		for i, u := range users {
			if u.Username == "" {
				return fmt.Errorf("entry %d: username is empty", i)
			}
			if seen[u.Username] {
				return fmt.Errorf("entry %d: duplicate username %q", i, u.Username)
			}
			seen[u.Username] = true
			if u.Password == "" {
				return fmt.Errorf("user %q: password hash is empty", u.Username)
			}
		}
		return nil
	})
}

func cmdEditAlbum(args []string) {
	fs := flag.NewFlagSet("edit-album", flag.ExitOnError)
	filePath := fs.String("file", "", "path to album.json (required)")
	fs.Parse(args)

	if *filePath == "" {
		fmt.Fprintf(os.Stderr, "error: -file is required for album editing\n")
		os.Exit(1)
	}

	editAndValidate(*filePath, func(data []byte) error {
		var cfg albumConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		if cfg.Access != nil && cfg.Access.View != "" {
			if !validAccessModes[cfg.Access.View] {
				return fmt.Errorf("invalid access mode: %q (must be public, authenticated, or restricted)", cfg.Access.View)
			}
		}
		return nil
	})
}

// editAndValidate implements visudo-style editing: copy file to temp,
// open in $EDITOR, validate, and only save back if valid.
func editAndValidate(filePath string, validate func([]byte) error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		fmt.Fprintf(os.Stderr, "error: $EDITOR is not set\n")
		os.Exit(1)
	}

	// Read existing file (or start with empty for new files).
	original, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", filePath, err)
		os.Exit(1)
	}
	if original == nil {
		original = []byte("{}\n")
	}

	// Create temp file with same extension for editor syntax highlighting.
	tmpFile, err := os.CreateTemp("", "gollery-edit-*.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating temp file: %v\n", err)
		os.Exit(1)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(original); err != nil {
		tmpFile.Close()
		fmt.Fprintf(os.Stderr, "error writing temp file: %v\n", err)
		os.Exit(1)
	}
	tmpFile.Close()

	for {
		// Open editor.
		cmd := exec.Command(editor, tmpPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error running editor: %v\n", err)
			os.Exit(1)
		}

		// Read edited content.
		edited, err := os.ReadFile(tmpPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading temp file: %v\n", err)
			os.Exit(1)
		}

		// Validate.
		if err := validate(edited); err != nil {
			fmt.Fprintf(os.Stderr, "\nValidation error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Press Enter to re-edit, or Ctrl+C to abort: ")
			var discard string
			fmt.Scanln(&discard)
			continue
		}

		// Check if content changed.
		if string(edited) == string(original) {
			fmt.Println("No changes made.")
			return
		}

		// Write back to original file.
		perm := os.FileMode(0644)
		if strings.HasSuffix(filePath, "users.json") || strings.Contains(filePath, "users") {
			perm = 0600
		}
		if err := os.WriteFile(filePath, edited, perm); err != nil {
			fmt.Fprintf(os.Stderr, "error writing %s: %v\n", filePath, err)
			os.Exit(1)
		}
		fmt.Printf("Saved %s\n", filePath)
		return
	}
}

func splitCSV(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
