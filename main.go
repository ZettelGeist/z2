package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Tag struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"unique;not null"`
}

type Note struct {
	ID      uint   `gorm:"primarykey"`
	Title   *string
	Summary *string
	Body    *string
	Tags    []Tag `gorm:"many2many:note_tags;"`
}

var db *gorm.DB
var dbFile string

var rootCmd = &cobra.Command{
	Use:   "z2",
	Short: "z2 is a simple note-taking system",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initDB()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		dbClose()
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new note",
	Run:   createNote,
}

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		fmt.Println("Failed to connect to the database:", err)
		os.Exit(1)
	}
	db.AutoMigrate(&Tag{}, &Note{})
}

func dbClose() {
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("Failed to close the database:", err)
		os.Exit(1)
	}
	sqlDB.Close()
}

func init() {
	// Set defaults from environment variables
	defaultDBFile := os.Getenv("Z2_DB_FILE")
	if defaultDBFile == "" {
		defaultDBFile = "z2.db"
	}

	// Define flags
	rootCmd.PersistentFlags().StringVar(&dbFile, "db", defaultDBFile, "Database file")
	rootCmd.AddCommand(createCmd)

	// Define flags for create command
	createCmd.Flags().StringP("title", "t", "", "Title of the note")
	createCmd.Flags().StringP("summary", "s", "", "Summary of the note")
	createCmd.Flags().StringSliceP("tags", "g", []string{}, "Tags for the note")
	createCmd.Flags().StringP("body-file", "b", "", "Path to the Markdown file for the note body")
}

func createNote(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	summary, _ := cmd.Flags().GetString("summary")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	bodyFilePath, _ := cmd.Flags().GetString("body-file")

	if bodyFilePath == "" {
		fmt.Println("Error: --body-file is required")
		os.Exit(1)
	}

	body, err := ioutil.ReadFile(bodyFilePath)
	if err != nil {
		fmt.Printf("Error reading from body file: %s\n", err)
		os.Exit(1)
	}
	bodyStr := string(body)

	var noteTags []Tag
	for _, tagName := range tags {
		tag := Tag{Name: tagName}
		db.FirstOrCreate(&tag, Tag{Name: tagName})
		noteTags = append(noteTags, tag)
	}

	note := Note{
		Title:   &title,
		Summary: &summary,
		Body:    &bodyStr,
		Tags:    noteTags,
	}

	db.Create(&note)
	fmt.Println("Note created successfully!")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
