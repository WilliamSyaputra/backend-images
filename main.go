package main

import (
	"log"
	"os"
	"path/filepath"

	_ "william/backend/docs" // Import untuk Swagger

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
)

func main() {
	app := fiber.New()
	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,DELETE",
		AllowHeaders: "Content-Type",
	}))

	// Endpoint Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api/v1")

	api.Get("/images", getImages)            // Lihat daftar gambar
	api.Get("/images/:name", getImage)       // Lihat satu gambar
	api.Post("/upload", uploadImage)         // Upload gambar
	api.Post("/uploads", uploadImages)       // Upload banyak gambar
	api.Delete("/images/:name", deleteImage) // Hapus gambar

	// Buat folder jika belum ada
	if _, err := os.Stat("images"); os.IsNotExist(err) {
		os.Mkdir("images", os.ModePerm)
	}

	log.Fatal(app.Listen(":3001"))
}

// get list images
func getImages(c *fiber.Ctx) error {
	files, err := os.ReadDir("images")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membaca folder"})
	}

	var images []string
	for _, file := range files {
		images = append(images, file.Name())
	}

	return c.JSON(images)
}

// Get Image by name
func getImage(c *fiber.Ctx) error {
	name := c.Params("name")
	path := filepath.Join("images", name)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{"error": "Gambar tidak ditemukan"})
	}

	c.Set("Content-Disposition", "attachment; filename="+name)
	c.Set("Content-Type", "application/octet-stream")
	return c.SendFile(path)
}

// Upload Image
func uploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File tidak ditemukan"})
	}

	path := filepath.Join("images", file.Filename)
	if err := c.SaveFile(file, path); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file"})
	}

	return c.JSON(fiber.Map{"message": "Upload sukses", "filename": file.Filename})
}

// upload multiple
func uploadImages(c *fiber.Ctx) error {
	// Ambil semua file yang diunggah
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Gagal membaca form data"})
	}

	files := form.File["files"] // Mengambil semua file dari input name="files"

	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Tidak ada file yang diunggah"})
	}

	var uploadedFiles []string

	for _, file := range files {
		path := filepath.Join("images", file.Filename)

		if err := c.SaveFile(file, path); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan file", "file": file.Filename})
		}

		uploadedFiles = append(uploadedFiles, file.Filename)
	}

	// Berikan respons sukses dengan daftar file yang telah diunggah
	return c.JSON(fiber.Map{"message": "Upload sukses", "files": uploadedFiles})
}

// Delete Image by name
func deleteImage(c *fiber.Ctx) error {
	name := c.Params("name")
	path := filepath.Join("images", name)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{"error": "Gambar tidak ditemukan"})
	}

	if err := os.Remove(path); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus file"})
	}

	return c.JSON(fiber.Map{"message": "Gambar berhasil dihapus"})
}
