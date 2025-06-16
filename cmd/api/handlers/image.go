package handlers

import (
	"context"

	"fmt"

	types "github.com/crane-cloud/mira-new/cmd/image-builder"

	utils "github.com/crane-cloud/mira-new/internals/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/open-ug/conveyor/pkg/client"
	cTypes "github.com/open-ug/conveyor/pkg/types"
)

type ImageHandler struct {
	client *client.Client
}

func NewImageHandler(cl *client.Client) *ImageHandler {
	if cl == nil {
		cl = client.NewClient()
	}
	return &ImageHandler{
		client: cl,
	}
}

func (h *ImageHandler) GenerateImage(c *fiber.Ctx) error {

	// the request body should be a FormData
	// and should contain the following fields:
	// - name: string
	// - type: git | file
	// - branch: string
	// - repo: string
	// - gitusername: string | optional
	// - gitpassword: string | optional

	// the request should be a multipart/form-data
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}
	// get the fields from the form
	name := form.Value["name"]
	sourceType := form.Value["type"]

	if len(name) == 0 || len(sourceType) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	if sourceType[0] != "git" && sourceType[0] != "file" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid source type",
		})
	}

	var app types.ImageBuild
	app.Name = name[0]

	// get the file from the form and check if it is a valid file. then save it to the server
	if sourceType[0] == "file" {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid file format",
			})
		}

		// Generate a unique name for the file
		uniqueName := fmt.Sprintf("%s-%s-%s", utils.GenerateRandomString(6), app.Name, file.Filename)

		// save the file to the server
		err = c.SaveFile(file, "./uploads/"+uniqueName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save file",
			})
		}

		// get server url protocol and host
		serverURL := fmt.Sprintf("%s://%s", c.Protocol(), string(c.Context().URI().Host()))
		app.Spec.Source.BlobFile.Source = serverURL + "/uploads/" + uniqueName
		app.Spec.Source.Type = "file"
	} else if sourceType[0] == "git" {
		// get the git fields from the form
		repo := form.Value["repo"]
		//branch := form.Value["branch"]
		//gitUsername := form.Value["gitusername"]
		//gitPassword := form.Value["gitpassword"]
		//if len(repo) == 0 || len(branch) == 0 {
		//	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		//		"error": "Missing required fields",
		//	})
		//}
		app.Spec.Source.GitRepo.URL = repo[0]
		//app.Spec.Source.GitRepo.Branch = branch[0]
		//app.Spec.Source.GitRepo.Username = gitUsername[0]
		//app.Spec.Source.GitRepo.Password = gitPassword[0]
		app.Spec.Source.Type = "git"
	}

	resp, err := h.client.CreateResource(context.Background(), &cTypes.Resource{
		Name:     app.Name,
		Resource: "image-builder",
		Spec:     app.Spec,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create application",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Image generation started",
		"data": fiber.Map{
			"name":   resp.Name,
			"runid":  resp.RunID,
			"wspath": h.client.GetAPIURL() + "/drivers/streams/logs/mira/" + resp.RunID,
		},
	})

}
