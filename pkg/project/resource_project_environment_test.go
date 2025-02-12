package project

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-shared/test"
	"golang.org/x/exp/slices"
)

func TestAccProjectEnvironment(t *testing.T) {
	name := fmt.Sprintf("env-%s", randSeq(10))
	projectKey := fmt.Sprintf("project-%s", strings.ToLower(randSeq(2)))
	resourceName := fmt.Sprintf("project_environment.%s", name)

	params := map[string]any{
		"env_id":      name,
		"name":        name,
		"project_key": projectKey,
	}

	template := `
		resource "project" "{{ .project_key }}" {
			key          = "{{ .project_key }}"
			display_name = "{{ .project_key }}"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
		}

		resource "project_environment" "{{ .env_id }}" {
			name        = "{{ .name }}"
			project_key = project.{{ .project_key }}.key
		}
	`

	enviroment := test.ExecuteTemplate("TestAccProjectEnvironment", template, params)

	updateParams := map[string]any{
		"env_id":      name,
		"name":        fmt.Sprintf("env-%s", randSeq(10)),
		"project_key": projectKey,
	}

	enviromentUpdated := test.ExecuteTemplate("TestAccProjectEnvironment", template, updateParams)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		CheckDestroy: verifyDeleted(resourceName, func(id string, request *resty.Request) (*resty.Response, error) {
			resp, err := verifyEnvironment(projectKey, id, request)
			return resp, err
		}),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: enviroment,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", params["name"].(string)),
					resource.TestCheckResourceAttr(resourceName, "project_key", params["project_key"].(string)),
				),
			},
			{
				Config: enviromentUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updateParams["name"].(string)),
					resource.TestCheckResourceAttr(resourceName, "project_key", updateParams["project_key"].(string)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s:%s", projectKey, updateParams["name"]),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccProjectEnvironment_invalid_length(t *testing.T) {
	name := fmt.Sprintf("env-%s", randSeq(14))
	projectKey := fmt.Sprintf("project-%s", strings.ToLower(randSeq(6)))
	resourceName := fmt.Sprintf("project_environment.%s", name)

	params := map[string]any{
		"name":        name,
		"project_key": projectKey,
	}

	template := `
		resource "project" "{{ .project_key }}" {
			key          = "{{ .project_key }}"
			display_name = "{{ .project_key }}"
			admin_privileges {
				manage_members   = true
				manage_resources = true
				index_resources  = true
			}
		}

		resource "project_environment" "{{ .name }}" {
			name        = "{{ .name }}"
			project_key = project.{{ .project_key }}.key
		}
	`

	enviroment := test.ExecuteTemplate("TestAccProjectEnvironment", template, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		CheckDestroy: verifyDeleted(resourceName, func(id string, request *resty.Request) (*resty.Response, error) {
			resp, err := verifyEnvironment(projectKey, id, request)
			return resp, err
		}),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config:      enviroment,
				ExpectError: regexp.MustCompile(`.*combined length of project_key and name \(separated by '-'\) cannot exceed 32 characters.*`),
			},
		},
	})
}

func verifyEnvironment(projectKey, id string, request *resty.Request) (*resty.Response, error) {
	envs := []ProjectEnvironment{}

	resp, err := request.
		SetPathParam("projectKey", projectKey).
		SetResult(&envs).
		Get(projectEnvironmentUrl)
	if err != nil {
		return resp, err
	}

	envExists := slices.ContainsFunc(envs, func(e ProjectEnvironment) bool {
		return e.Name == fmt.Sprintf("%s-%s", projectKey, id)
	})

	if !envExists {
		return resp, fmt.Errorf("environment %s does not exist", id)
	}

	return resp, nil
}
