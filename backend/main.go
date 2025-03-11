package main

import (
	"database/sql"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error
	dsn := "root:root@tcp(127.0.0.1:3306)/course_db"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	if err := db.Ping(); err != nil {
		log.Fatal("Database unreachable:", err)
	}
}

// CourseResponse now has Duration as int.
type CourseResponse struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	Image       string `json:"image"` // base64-encoded image string with proper MIME type
}

// addCourse accepts multipart/form-data and stores the image and PDF as BLOBs.
func addCourse(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	durationStr := c.PostForm("duration")

	// Validate description: must not exceed 100 characters.
	if len(description) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description must be 100 characters or less"})
		return
	}

	// Convert duration from string to int.
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration, must be an integer"})
		return
	}

	// Process image file.
	imageFile, _, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload image"})
		return
	}
	defer imageFile.Close()
	imageBytes, err := io.ReadAll(imageFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading image file"})
		return
	}

	// Process PDF file.
	pdfFile, _, err := c.Request.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload pdf"})
		return
	}
	defer pdfFile.Close()
	pdfBytes, err := io.ReadAll(pdfFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading pdf file"})
		return
	}

	// Insert into the database (duration is now an int).
	query := "INSERT INTO courses (title, description, duration, image, pdf) VALUES (?, ?, ?, ?, ?)"
	_, err = db.Exec(query, title, description, duration, imageBytes, pdfBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course added successfully"})
}

// updateCourse updates an existing course record.
// If new files (image and/or pdf) are provided, they replace the old ones; otherwise, the current ones are retained.
func updateCourse(c *gin.Context) {
	id := c.Param("id")
	title := c.PostForm("title")
	description := c.PostForm("description")
	durationStr := c.PostForm("duration")

	// Validate description length.
	if len(description) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description must be 100 characters or less"})
		return
	}

	// Convert duration to int.
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration, must be an integer"})
		return
	}

	// Process image file.
	var imageBytes []byte
	imageFile, _, err := c.Request.FormFile("image")
	if err == nil {
		defer imageFile.Close()
		imageBytes, err = io.ReadAll(imageFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading image file"})
			return
		}
	} else {
		// No new image provided: fetch the current image from the DB.
		row := db.QueryRow("SELECT image FROM courses WHERE id = ?", id)
		if err := row.Scan(&imageBytes); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
			return
		}
	}

	// Process PDF file.
	var pdfBytes []byte
	pdfFile, _, err := c.Request.FormFile("pdf")
	if err == nil {
		defer pdfFile.Close()
		pdfBytes, err = io.ReadAll(pdfFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading pdf file"})
			return
		}
	} else {
		// No new pdf provided: fetch the current pdf from the DB.
		row := db.QueryRow("SELECT pdf FROM courses WHERE id = ?", id)
		if err := row.Scan(&pdfBytes); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
			return
		}
	}

	// Update the course record in the database.
	query := "UPDATE courses SET title = ?, description = ?, duration = ?, image = ?, pdf = ? WHERE id = ?"
	_, err = db.Exec(query, title, description, duration, imageBytes, pdfBytes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update course"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Course updated successfully"})
}

// getCourses fetches only id, title, description, duration, and image (the PDF is not returned).
func getCourses(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, description, duration, image FROM courses")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch courses"})
		return
	}
	defer rows.Close()

	var courses []CourseResponse
	for rows.Next() {
		var course CourseResponse
		var imageBytes []byte
		if err := rows.Scan(&course.ID, &course.Title, &course.Description, &course.Duration, &imageBytes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning course data"})
			return
		}
		// Dynamically detect the MIME type for the image.
		mimeType := http.DetectContentType(imageBytes)
		course.Image = "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(imageBytes)
		courses = append(courses, course)
	}
	c.JSON(http.StatusOK, courses)
}

func main() {
	r := gin.Default()

	// Endpoints for adding, updating, and fetching courses.
	r.POST("/add-course", addCourse)
	r.PUT("/update-course/:id", updateCourse)
	r.GET("/courses", getCourses)

	r.Run(":8080")
}
