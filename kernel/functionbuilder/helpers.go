package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// validateRequest validates the incoming request for the upload.
func validateRequest(w http.ResponseWriter, r *http.Request) bool {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Println("Invalid request method")
		return false
	}

	// Parse the form with a 10 MB limit
	log.Println("Parsing multipart form")
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		log.Printf("Failed to parse form: %v\n", err)
		return false
	}

	// Retrieve the image name
	if r.FormValue("image_name") == "" {
		http.Error(w, "Image name is required", http.StatusBadRequest)
		log.Printf("Image name is required: %v\n", err)
		return false
	}
	return true
}

// createBuildDir creates a temporary directory for processing.
func createBuildDir(w http.ResponseWriter, imageName string) (string, bool) {
	log.Println("Creating directory for processing")
	buildDir, err := os.MkdirTemp("", imageName)
	if err != nil {
		http.Error(w, "Failed to create temp directory", http.StatusInternalServerError)
		log.Printf("Failed to create temp directory: %v\n", err)
		return "", false
	}
	log.Printf("Temporary directory created: %s\n", buildDir)
	return buildDir, true
}

// saveUploadedFile saves the uploaded file to the temporary directory.
func saveUploadedFile(w http.ResponseWriter, r *http.Request, buildDir string) bool {
	// Parse the multipart form with a 10 MB limit
	log.Println("Parsing multipart form")
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		log.Printf("Failed to parse form: %v\n", err)
		return false
	}

	// Retrieve the uploaded files
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		log.Printf("No files uploaded")
		return false
	}

	for _, fileHeader := range files {
		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to retrieve files", http.StatusBadRequest)
			log.Printf("Failed to retrieve file: %v\n", err)
			return false
		}
		defer file.Close()

		log.Printf("File %s retrieved successfully", fileHeader.Filename)

		// Create the destination file
		destPath := filepath.Join(buildDir, fileHeader.Filename)
		out, err := os.Create(destPath)
		if err != nil {
			http.Error(w, "Failed to create destination file", http.StatusInternalServerError)
			log.Printf("Failed to create destination file: %v\n", err)
			return false
		}
		defer out.Close()
		log.Printf("Destination file %s created successfully", destPath)

		// Copy the uploaded file to the destination
		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Failed to copy file", http.StatusInternalServerError)
			log.Printf("Failed to copy file: %v\n", err)
			return false
		}
		log.Printf("File %s copied successfully", fileHeader.Filename)
	}
	log.Println("All files saved successfully")
	return true
}

func buildImage(w http.ResponseWriter, imageName string, buildDir string) bool {
	// Build the Podman image with the custom image name
	log.Printf("Building image %s", imageName)
	log.Printf("Command: buildah bud --isolation=chroot -t %s %s", imageName, buildDir)
	cmd := exec.Command("buildah", "bud", "--isolation=chroot", "-t", imageName, buildDir)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out // Capture standard error as well

	err := cmd.Run()
	if err != nil {
		http.Error(w, "Failed to build image", http.StatusInternalServerError)
		log.Printf("Failed to build image: %v\n", err)
		log.Println(out.String())
		return false
	}
	log.Printf("Image %s built successfully", imageName)
	return true
}

func tagImage(w http.ResponseWriter, imageName string) (string, bool) {
	taggedImage := imageName + ":latest"
	log.Printf("Taging image %s", imageName)
	log.Printf("Command: buildah tag %s %s", imageName, taggedImage)
	cmd := exec.Command("buildah", "tag", imageName, taggedImage)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out // Capture standard error as well

	err := cmd.Run()
	if err != nil {
		http.Error(w, "Failed to Tag image", http.StatusInternalServerError)
		log.Printf("Failed to Tag image: %v\n", err)
		log.Println(out.String())
		return taggedImage, false
	}
	log.Printf("Image %s tagged", taggedImage)
	return taggedImage, true
}

func loginDockerIo(w http.ResponseWriter) bool {
	// Load username and password from environment variables
	username := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")

	if username == "" || password == "" {
		http.Error(w, "Credentials not provided", http.StatusInternalServerError)
		log.Println("Username or password environment variable is not set")
		return false
	}

	log.Println("Login Docker IO")
	log.Printf("Command: buildah login -u %s -p %s docker.io", username, password)
	cmd := exec.Command("buildah", "login", "-u", username, "-p", password, "docker.io")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out // Capture standard error as well

	err := cmd.Run()
	if err != nil {
		http.Error(w, "Failed to login", http.StatusInternalServerError)
		log.Printf("Failed to login: %v\n", err)
		log.Println(out.String())
		return false
	}
	log.Println("Logged")
	return true
}

func pushImage(w http.ResponseWriter, taggedImage string) bool {
	log.Println("Push image")
	log.Printf("Command: buildah push %s % s", taggedImage, taggedImage)
	cmd := exec.Command("buildah", "push", taggedImage, "docker.io/quangnguyen2017/"+taggedImage)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out // Capture standard error as well

	err := cmd.Run()
	if err != nil {
		http.Error(w, "Failed to push image", http.StatusInternalServerError)
		log.Printf("Failed to push image: %v\n", err)
		log.Println(out.String())
		return false
	}
	log.Printf("Image %s pushed", taggedImage)
	return true
}
