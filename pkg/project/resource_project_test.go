package project

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-shared/test"
)

func verifyProject(id string, request *resty.Request) (*resty.Response, error) {
	return request.Head(projectsUrl + id)
}

func getRandomMaxStorageSize() int {
	randomMaxStorage := rand.Intn(maxStorageInGibibytes)
	if randomMaxStorage == 0 {
		randomMaxStorage = 1
	}

	return randomMaxStorage
}

func makeInvalidProjectKeyTestCase(invalidProjectKey string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("tftestprojects%s", randSeq(10))
	resourceName := fmt.Sprintf("project.%s", name)

	params := map[string]interface{}{
		"max_storage_in_gibibytes":   getRandomMaxStorageSize(),
		"block_deployments_on_limit": test.RandBool(),
		"email_notification":         test.RandBool(),
		"manage_members":             test.RandBool(),
		"manage_resources":           test.RandBool(),
		"index_resources":            test.RandBool(),
		"name":                       name,
		"project_key":                invalidProjectKey, //strings.ToLower(randSeq(20)),
	}
	project := test.ExecuteTemplate("TestAccProjects", `
		resource "project" "{{ .name }}" {
			key = "{{ .project_key }}"
			display_name = "{{ .name }}"
			description = "test description"
			admin_privileges {
				manage_members = {{ .manage_members }}
				manage_resources = {{ .manage_resources }}
				index_resources = {{ .index_resources }}
			}
			max_storage_in_gibibytes = {{ .max_storage_in_gibibytes }}
			block_deployments_on_limit = {{ .block_deployments_on_limit }}
			email_notification = {{ .email_notification }}
		}
	`, params)

	return t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, verifyProject),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config:      project,
				ExpectError: regexp.MustCompile(".*project_key must be 2 - 32 lowercase alphanumeric and hyphen characters.*"),
			},
		},
	}
}

type testCase struct {
	Name  string
	Value string
}

func TestAccProjectInvalidProjectKey(t *testing.T) {
	invalidProjectKeys := []testCase{
		{
			Name:  "TooShort",
			Value: strings.ToLower(randSeq(1)),
		},
		{
			Name:  "TooLong",
			Value: strings.ToLower(randSeq(33)),
		},
		{
			Name:  "HasUppercase",
			Value: randSeq(8),
		},
	}

	for _, invalidProjectKey := range invalidProjectKeys {
		t.Run(invalidProjectKey.Name, func(t *testing.T) {
			resource.Test(makeInvalidProjectKeyTestCase(invalidProjectKey.Value, t))
		})
	}
}

func testProjectConfig(name, key string) string {
	params := map[string]interface{}{
		"max_storage_in_gibibytes":   getRandomMaxStorageSize(),
		"block_deployments_on_limit": test.RandBool(),
		"email_notification":         test.RandBool(),
		"manage_members":             test.RandBool(),
		"manage_resources":           test.RandBool(),
		"index_resources":            test.RandBool(),
		"name":                       name,
		"project_key":                key,
	}
	return test.ExecuteTemplate("TestAccProjects", `
		resource "project" "{{ .name }}" {
			key = "{{ .project_key }}"
			display_name = "{{ .name }}"
			description = "test description"
			admin_privileges {
				manage_members = {{ .manage_members }}
				manage_resources = {{ .manage_resources }}
				index_resources = {{ .index_resources }}
			}
			max_storage_in_gibibytes = {{ .max_storage_in_gibibytes }}
			block_deployments_on_limit = {{ .block_deployments_on_limit }}
			email_notification = {{ .email_notification }}
		}
	`, params)
}

func TestAccProjectInvalidMaxStorage(t *testing.T) {
	invalidMaxStorages := []struct {
		Name       string
		Value      int64
		ErrorRegex string
	}{
		{
			Name:       "Invalid",
			Value:      -2,
			ErrorRegex: `.*expected max_storage_in_gibibytes to be one of \[-1\], got -2.*`,
		},
		{
			Name:       "TooSmall",
			Value:      0,
			ErrorRegex: `.*expected max_storage_in_gibibytes to be in the range \(1 - 8589934591\), got 0.*`,
		},
		{
			Name:       "TooLarge",
			Value:      8589934592,
			ErrorRegex: `.*expected max_storage_in_gibibytes to be in the range \(1 - 8589934591\), got 8589934592.*`,
		},
	}

	for _, invalidMaxStorage := range invalidMaxStorages {
		t.Run(invalidMaxStorage.Name, func(t *testing.T) {
			resource.Test(makeInvalidMaxStorageTestCase(invalidMaxStorage.Value, invalidMaxStorage.ErrorRegex, t))
		})
	}
}

func makeInvalidMaxStorageTestCase(invalidMaxStorage int64, errorRegex string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("tftestprojects%s", randSeq(10))
	resourceName := fmt.Sprintf("project.%s", name)

	params := map[string]interface{}{
		"max_storage_in_gibibytes":   invalidMaxStorage,
		"block_deployments_on_limit": test.RandBool(),
		"email_notification":         test.RandBool(),
		"manage_members":             test.RandBool(),
		"manage_resources":           test.RandBool(),
		"index_resources":            test.RandBool(),
		"name":                       name,
		"project_key":                strings.ToLower(randSeq(20)),
	}
	project := test.ExecuteTemplate("TestAccProjects", `
		resource "project" "{{ .name }}" {
			key = "{{ .project_key }}"
			display_name = "{{ .name }}"
			description = "test description"
			admin_privileges {
				manage_members = {{ .manage_members }}
				manage_resources = {{ .manage_resources }}
				index_resources = {{ .index_resources }}
			}
			max_storage_in_gibibytes = {{ .max_storage_in_gibibytes }}
			block_deployments_on_limit = {{ .block_deployments_on_limit }}
			email_notification = {{ .email_notification }}
		}
	`, params)

	return t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, verifyProject),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config:      project,
				ExpectError: regexp.MustCompile(errorRegex),
			},
		},
	}
}

func TestAccProjectInvalidDisplayName(t *testing.T) {
	name := fmt.Sprintf("invalidtestprojects%s", randSeq(20))
	resourceName := fmt.Sprintf("project.%s", name)
	project := testProjectConfig(name, strings.ToLower(randSeq(6)))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, verifyProject),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config:      project,
				ExpectError: regexp.MustCompile(`.*string must be less than or equal 32 characters long.*`),
			},
		},
	})
}

func TestAccProjectUpdateKey(t *testing.T) {
	name := fmt.Sprintf("testprojects%s", randSeq(20))
	resourceName := fmt.Sprintf("project.%s", name)
	key1 := strings.ToLower(randSeq(6))
	config := testProjectConfig(name, key1)

	key2 := strings.ToLower(randSeq(6))
	configWithNewKey := testProjectConfig(name, key2)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, verifyProject),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key1),
					resource.TestCheckResourceAttr(resourceName, "display_name", name),
				),
			},
			{
				Config: configWithNewKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key2),
					resource.TestCheckResourceAttr(resourceName, "display_name", name),
				),
			},
		},
	})
}

func TestAccProject_full(t *testing.T) {
	name := fmt.Sprintf("tftestprojects%s", randSeq(10))
	resourceName := fmt.Sprintf("project.%s", name)

	username1 := "user1"
	email1 := username1 + "@tempurl.org"
	username2 := "user2"
	email2 := username2 + "@tempurl.org"
	group1 := "group1"
	group2 := "group2"
	repo1 := fmt.Sprintf("repo%d", test.RandomInt())
	repo2 := fmt.Sprintf("repo%d", test.RandomInt())

	params := map[string]interface{}{
		"max_storage_in_gibibytes":   getRandomMaxStorageSize(),
		"block_deployments_on_limit": test.RandBool(),
		"email_notification":         test.RandBool(),
		"manage_members":             test.RandBool(),
		"manage_resources":           test.RandBool(),
		"index_resources":            test.RandBool(),
		"name":                       name,
		"project_key":                strings.ToLower(randSeq(6)),
		"username1":                  username1,
		"username2":                  username2,
		"email1":                     email1,
		"email2":                     email2,
		"group1":                     group1,
		"group2":                     group2,
		"repo1":                      repo1,
		"repo2":                      repo2,
	}

	template := `
		resource "artifactory_managed_user" "{{ .username1 }}" {
			name     = "{{ .username1 }}"
			email    = "{{ .email1 }}"
			password = "Password1!"
			admin    = false
		}

		resource "artifactory_managed_user" "{{ .username2 }}" {
			name     = "{{ .username2 }}"
			email    = "{{ .email2 }}"
			password = "Password1!"
			admin    = false
		}

		resource "artifactory_group" "{{ .group1 }}" {
			name = "{{ .group1 }}"
		}

		resource "artifactory_group" "{{ .group2 }}" {
			name = "{{ .group2 }}"
		}

		resource "artifactory_local_generic_repository" "{{ .repo1 }}" {
			key = "{{ .repo1 }}"

			lifecycle {
				ignore_changes = ["project_key"]
			}
		}

		resource "artifactory_local_generic_repository" "{{ .repo2 }}" {
			key = "{{ .repo2 }}"

			lifecycle {
				ignore_changes = ["project_key"]
			}
		}

		resource "project" "{{ .name }}" {
			key = "{{ .project_key }}"
			display_name = "{{ .name }}"
			description = "test description"
			admin_privileges {
				manage_members = {{ .manage_members }}
				manage_resources = {{ .manage_resources }}
				index_resources = {{ .index_resources }}
			}
			max_storage_in_gibibytes = {{ .max_storage_in_gibibytes }}
			block_deployments_on_limit = {{ .block_deployments_on_limit }}
			email_notification = {{ .email_notification }}

			member {
				name  = artifactory_managed_user.{{ .username1 }}.name
				roles = ["Developer","Project Admin"]
			}

			member {
				name  = artifactory_managed_user.{{ .username2 }}.name
				roles = ["Developer"]
			}

			group {
				name  = artifactory_group.{{ .group1 }}.name
				roles = ["qa"]
			}

			group {
				name  = artifactory_group.{{ .group2 }}.name
				roles = ["Release Manager"]
			}

			role {
				name         = "qa"
				description  = "QA role"
				type         = "CUSTOM"
				environments = ["DEV"]
				actions      = ["READ_REPOSITORY","READ_RELEASE_BUNDLE", "READ_BUILD", "READ_SOURCES_PIPELINE", "READ_INTEGRATIONS_PIPELINE", "READ_POOLS_PIPELINE", "TRIGGER_PIPELINE"]
			}

			role {
				name         = "devop"
				description  = "DevOp role"
				type         = "CUSTOM"
				environments = ["DEV", "PROD"]
				actions      = ["READ_REPOSITORY", "ANNOTATE_REPOSITORY", "DEPLOY_CACHE_REPOSITORY", "DELETE_OVERWRITE_REPOSITORY", "TRIGGER_PIPELINE", "READ_INTEGRATIONS_PIPELINE", "READ_POOLS_PIPELINE", "MANAGE_INTEGRATIONS_PIPELINE", "MANAGE_SOURCES_PIPELINE", "MANAGE_POOLS_PIPELINE", "READ_BUILD", "ANNOTATE_BUILD", "DEPLOY_BUILD", "DELETE_BUILD",]
			}

			repos = [
				artifactory_local_generic_repository.{{ .repo1 }}.key,
				artifactory_local_generic_repository.{{ .repo2 }}.key,
			]
		}
	`

	project := test.ExecuteTemplate("TestAccProjects", template, params)

	updateParams := map[string]interface{}{
		"max_storage_in_gibibytes":   params["max_storage_in_gibibytes"],
		"block_deployments_on_limit": !params["block_deployments_on_limit"].(bool),
		"email_notification":         !params["email_notification"].(bool),
		"manage_members":             !params["manage_members"].(bool),
		"manage_resources":           !params["manage_resources"].(bool),
		"index_resources":            !params["index_resources"].(bool),
		"name":                       params["name"],
		"project_key":                params["project_key"],
		"username1":                  params["username1"],
		"username2":                  params["username2"],
		"email1":                     params["email1"],
		"email2":                     params["email2"],
		"group1":                     params["group1"],
		"group2":                     params["group2"],
		"repo1":                      params["repo1"],
		"repo2":                      params["repo2"],
	}
	projectUpdated := test.ExecuteTemplate("TestAccProjects", template, updateParams)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, verifyProject),
		ProviderFactories: testAccProviders(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"artifactory": {
				Source:            "jfrog/artifactory",
				VersionConstraint: "10.1.3",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: project,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", fmt.Sprintf("%s", params["project_key"])),
					resource.TestCheckResourceAttr(resourceName, "display_name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "max_storage_in_gibibytes", fmt.Sprintf("%d", params["max_storage_in_gibibytes"])),
					resource.TestCheckResourceAttr(resourceName, "block_deployments_on_limit", fmt.Sprintf("%t", params["block_deployments_on_limit"])),
					resource.TestCheckResourceAttr(resourceName, "email_notification", fmt.Sprintf("%t", params["email_notification"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.manage_members", fmt.Sprintf("%t", params["manage_members"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.manage_resources", fmt.Sprintf("%t", params["manage_resources"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.index_resources", fmt.Sprintf("%t", params["index_resources"])),
					resource.TestCheckResourceAttr(resourceName, "use_project_role_resource", "false"),
					resource.TestCheckResourceAttr(resourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "member.0.name", username1),
					resource.TestCheckResourceAttr(resourceName, "member.0.roles.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "member.0.roles.0", "Developer"),
					resource.TestCheckResourceAttr(resourceName, "member.0.roles.1", "Project Admin"),
					resource.TestCheckResourceAttr(resourceName, "member.1.name", username2),
					resource.TestCheckResourceAttr(resourceName, "member.1.roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "member.1.roles.0", "Developer"),
					resource.TestCheckResourceAttr(resourceName, "group.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "group.0.name", group1),
					resource.TestCheckResourceAttr(resourceName, "group.0.roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "group.0.roles.0", "qa"),
					resource.TestCheckResourceAttr(resourceName, "group.1.name", group2),
					resource.TestCheckResourceAttr(resourceName, "group.1.roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "group.1.roles.0", "Release Manager"),
					resource.TestCheckResourceAttr(resourceName, "repos.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "repos.*", repo1),
					resource.TestCheckTypeSetElemAttr(resourceName, "repos.*", repo2),
				),
			},
			{
				Config: projectUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", fmt.Sprintf("%s", updateParams["project_key"])),
					resource.TestCheckResourceAttr(resourceName, "display_name", name),
					resource.TestCheckResourceAttr(resourceName, "max_storage_in_gibibytes", fmt.Sprintf("%d", updateParams["max_storage_in_gibibytes"])),
					resource.TestCheckResourceAttr(resourceName, "block_deployments_on_limit", fmt.Sprintf("%t", updateParams["block_deployments_on_limit"])),
					resource.TestCheckResourceAttr(resourceName, "email_notification", fmt.Sprintf("%t", updateParams["email_notification"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.manage_members", fmt.Sprintf("%t", updateParams["manage_members"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.manage_resources", fmt.Sprintf("%t", updateParams["manage_resources"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.index_resources", fmt.Sprintf("%t", updateParams["index_resources"])),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"use_project_role_resource"},
			},
		},
	})
}

func TestAccProject_migrate_schema(t *testing.T) {
	name := fmt.Sprintf("tftestprojects%s", randSeq(10))
	resourceName := fmt.Sprintf("project.%s", name)

	params := map[string]interface{}{
		"max_storage_in_gibibytes":   getRandomMaxStorageSize(),
		"block_deployments_on_limit": test.RandBool(),
		"email_notification":         test.RandBool(),
		"manage_members":             test.RandBool(),
		"manage_resources":           test.RandBool(),
		"index_resources":            test.RandBool(),
		"name":                       name,
		"project_key":                strings.ToLower(randSeq(6)),
	}

	template := `
		resource "project" "{{ .name }}" {
			key = "{{ .project_key }}"
			display_name = "{{ .name }}"
			description = "test description"
			admin_privileges {
				manage_members = {{ .manage_members }}
				manage_resources = {{ .manage_resources }}
				index_resources = {{ .index_resources }}
			}
			max_storage_in_gibibytes = {{ .max_storage_in_gibibytes }}
			block_deployments_on_limit = {{ .block_deployments_on_limit }}
			email_notification = {{ .email_notification }}

			role {
				name         = "qa"
				description  = "QA role"
				type         = "CUSTOM"
				environments = ["DEV"]
				actions      = ["READ_REPOSITORY","READ_RELEASE_BUNDLE", "READ_BUILD", "READ_SOURCES_PIPELINE", "READ_INTEGRATIONS_PIPELINE", "READ_POOLS_PIPELINE", "TRIGGER_PIPELINE"]
			}

			role {
				name         = "devop"
				description  = "DevOp role"
				type         = "CUSTOM"
				environments = ["DEV", "PROD"]
				actions      = ["READ_REPOSITORY", "ANNOTATE_REPOSITORY", "DEPLOY_CACHE_REPOSITORY", "DELETE_OVERWRITE_REPOSITORY", "TRIGGER_PIPELINE", "READ_INTEGRATIONS_PIPELINE", "READ_POOLS_PIPELINE", "MANAGE_INTEGRATIONS_PIPELINE", "MANAGE_SOURCES_PIPELINE", "MANAGE_POOLS_PIPELINE", "READ_BUILD", "ANNOTATE_BUILD", "DEPLOY_BUILD", "DELETE_BUILD",]
			}
		}
	`

	config := test.ExecuteTemplate("TestAccProject", template, params)

	updatedTemplate := `
		resource "project" "{{ .name }}" {
			key = "{{ .project_key }}"
			display_name = "{{ .name }}"
			description = "test description"
			admin_privileges {
				manage_members = {{ .manage_members }}
				manage_resources = {{ .manage_resources }}
				index_resources = {{ .index_resources }}
			}
			max_storage_in_gibibytes = {{ .max_storage_in_gibibytes }}
			block_deployments_on_limit = {{ .block_deployments_on_limit }}
			email_notification = {{ .email_notification }}
			use_project_role_resource = true
		}
	`

	updateParams := map[string]interface{}{
		"max_storage_in_gibibytes":   params["max_storage_in_gibibytes"],
		"block_deployments_on_limit": !params["block_deployments_on_limit"].(bool),
		"email_notification":         !params["email_notification"].(bool),
		"manage_members":             !params["manage_members"].(bool),
		"manage_resources":           !params["manage_resources"].(bool),
		"index_resources":            !params["index_resources"].(bool),
		"name":                       params["name"],
		"project_key":                params["project_key"],
	}
	updatedConfig := test.ExecuteTemplate("TestAccProject", updatedTemplate, updateParams)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: verifyDeleted(resourceName, verifyProject),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"project": {
						VersionConstraint: "1.2.1",
						Source:            "registry.terraform.io/jfrog/project",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", params["project_key"].(string)),
					resource.TestCheckResourceAttr(resourceName, "display_name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "max_storage_in_gibibytes", fmt.Sprintf("%d", params["max_storage_in_gibibytes"])),
					resource.TestCheckResourceAttr(resourceName, "block_deployments_on_limit", fmt.Sprintf("%t", params["block_deployments_on_limit"])),
					resource.TestCheckResourceAttr(resourceName, "email_notification", fmt.Sprintf("%t", params["email_notification"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.manage_members", fmt.Sprintf("%t", params["manage_members"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.manage_resources", fmt.Sprintf("%t", params["manage_resources"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.index_resources", fmt.Sprintf("%t", params["index_resources"])),
					resource.TestCheckNoResourceAttr(resourceName, "use_project_role_resource"),
					resource.TestCheckResourceAttr(resourceName, "role.#", "2"),
				),
			},
			{
				ProviderFactories: testAccProviders(),
				Config:            config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", params["project_key"].(string)),
					resource.TestCheckResourceAttr(resourceName, "display_name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "max_storage_in_gibibytes", fmt.Sprintf("%d", params["max_storage_in_gibibytes"])),
					resource.TestCheckResourceAttr(resourceName, "block_deployments_on_limit", fmt.Sprintf("%t", params["block_deployments_on_limit"])),
					resource.TestCheckResourceAttr(resourceName, "email_notification", fmt.Sprintf("%t", params["email_notification"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.manage_members", fmt.Sprintf("%t", params["manage_members"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.manage_resources", fmt.Sprintf("%t", params["manage_resources"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.index_resources", fmt.Sprintf("%t", params["index_resources"])),
					resource.TestCheckResourceAttr(resourceName, "role.#", "2"),
				),
			},
			{
				ProviderFactories: testAccProviders(),
				Config:            updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", params["project_key"].(string)),
					resource.TestCheckResourceAttr(resourceName, "display_name", name),
					resource.TestCheckResourceAttr(resourceName, "max_storage_in_gibibytes", fmt.Sprintf("%d", updateParams["max_storage_in_gibibytes"])),
					resource.TestCheckResourceAttr(resourceName, "block_deployments_on_limit", fmt.Sprintf("%t", updateParams["block_deployments_on_limit"])),
					resource.TestCheckResourceAttr(resourceName, "email_notification", fmt.Sprintf("%t", updateParams["email_notification"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.manage_members", fmt.Sprintf("%t", updateParams["manage_members"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.manage_resources", fmt.Sprintf("%t", updateParams["manage_resources"])),
					resource.TestCheckResourceAttr(resourceName, "admin_privileges.0.index_resources", fmt.Sprintf("%t", updateParams["index_resources"])),
					resource.TestCheckResourceAttr(resourceName, "use_project_role_resource", "true"),
					resource.TestCheckResourceAttr(resourceName, "role.#", "0"),
				),
			},
		},
	})
}
