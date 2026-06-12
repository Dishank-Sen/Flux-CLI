package debug

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Dishank-Sen/Flux-CLI/cli/utils"
)

func TestDebug(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Home directory:", home)
}

func PromptEmail() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(email), nil
}

func TestPromptEmail(t *testing.T) {
	input := "user1@gmail.com\n"

	r, w, _ := os.Pipe()

	// write fake user input
	w.Write([]byte(input))
	w.Close()

	// replace stdin
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	email, err := PromptEmail()
	if err != nil {
		t.Fatal(err)
	}

	if email != "user1@gmail.com" {
		t.Fatalf("expected user1@gmail.com got %s", email)
	}
}

func TestFileTree(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		t.Fatal(err)
	}

	if err := utils.CreateFileTree(context.Background()); err != nil {
		t.Fatal(err)
	}

	fmt.Println("file tree created successfully!")
}
