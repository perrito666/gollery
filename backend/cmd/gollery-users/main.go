// Command gollery-users manages the users.json file used by gollery's
// static auth provider. It supports listing, adding, removing users,
// changing passwords, and toggling admin status.
//
// Usage:
//
//	gollery-users -file users.json list
//	gollery-users -file users.json add -username alice -password secret
//	gollery-users -file users.json add -username alice -password secret -admin -groups editors,viewers
//	gollery-users -file users.json remove -username alice
//	gollery-users -file users.json passwd -username alice -password newpass
//	gollery-users -file users.json set-admin -username alice -admin
//	gollery-users -file users.json set-admin -username alice -admin=false
//	gollery-users -file users.json set-groups -username alice -groups editors,viewers
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: gollery-users -file <users.json> <command> [flags]

Commands:
  list                              List all users
  add       -username U -password P Add a new user (optional: -admin, -groups g1,g2)
  remove    -username U             Remove a user
  passwd    -username U -password P Change a user's password
  set-admin -username U -admin      Set or clear admin flag (-admin or -admin=false)
  set-groups -username U -groups G  Set user groups (comma-separated, empty to clear)
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
