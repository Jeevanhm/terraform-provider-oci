// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/oracle/oci-go-sdk/common"
	oci_identity "github.com/oracle/oci-go-sdk/identity"
)

const (
	PolicyRequiredOnlyResource = PolicyResourceDependencies + `
resource "oci_identity_policy" "test_policy" {
	#Required
	compartment_id = "${var.tenancy_ocid}"
	description = "${var.policy_description}"
	name = "${var.policy_name}"
	statements = ["Allow group ${oci_identity_group.t.name} to read instances in compartment ${oci_identity_compartment.t.name}"]
}
`

	PolicyResourceConfig = PolicyResourceDependencies + `
resource "oci_identity_policy" "test_policy" {
	#Required
	compartment_id = "${var.tenancy_ocid}"
	description = "${var.policy_description}"
	name = "${var.policy_name}"
	statements = ["Allow group ${oci_identity_group.t.name} to read instances in compartment ${oci_identity_compartment.t.name}"]

	#Optional
	defined_tags = "${map("${oci_identity_tag_namespace.tag-namespace1.name}.${oci_identity_tag.tag1.name}", "${var.policy_defined_tags_value}")}"
	freeform_tags = "${var.policy_freeform_tags}"
	version_date = "${var.policy_version_date}"
}
`
	PolicyPropertyVariables = `
variable "policy_defined_tags_value" { default = "value" }
variable "policy_description" { default = "Policy for users who need to launch instances, attach volumes, manage images" }
variable "policy_freeform_tags" { default = {"Department"= "Finance"} }
variable "policy_name" { default = "LaunchInstances" }
variable "policy_version_date" { default = "" }

`
	PolicyResourceDependencies = DefinedTagsDependencies + `
resource "oci_identity_compartment" "t" {
	name = "Network"
	description = "For network components"
}

resource "oci_identity_group" "t" {
	#Required
	compartment_id = "${var.tenancy_ocid}"
	description = "group for policy test"
	name = "GroupName"
}
`
)

func TestIdentityPolicyResource_basic(t *testing.T) {
	provider := testAccProvider
	config := testProviderConfig()

	compartmentId := getEnvSettingWithBlankDefault("compartment_ocid")
	compartmentIdVariableStr := fmt.Sprintf("variable \"compartment_id\" { default = \"%s\" }\n", compartmentId)
	tenancyId := getEnvSettingWithBlankDefault("tenancy_ocid")

	resourceName := "oci_identity_policy.test_policy"
	datasourceName := "data.oci_identity_policies.test_policies"

	var resId, resId2 string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: map[string]terraform.ResourceProvider{
			"oci": provider,
		},
		CheckDestroy: testAccCheckIdentityPolicyDestroy,
		Steps: []resource.TestStep{
			// verify create
			{
				Config: config + PolicyPropertyVariables + compartmentIdVariableStr + PolicyRequiredOnlyResource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", tenancyId),
					resource.TestCheckResourceAttr(resourceName, "description", "Policy for users who need to launch instances, attach volumes, manage images"),
					resource.TestCheckResourceAttr(resourceName, "name", "LaunchInstances"),
					resource.TestCheckResourceAttr(resourceName, "statements.#", "1"),

					func(s *terraform.State) (err error) {
						resId, err = fromInstanceState(s, resourceName, "id")
						return err
					},
				),
			},

			// delete before next create
			{
				Config: config + compartmentIdVariableStr + PolicyResourceDependencies,
			},
			// verify create with optionals
			{
				Config: config + PolicyPropertyVariables + compartmentIdVariableStr + PolicyResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", tenancyId),
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "Policy for users who need to launch instances, attach volumes, manage images"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "LaunchInstances"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "statements.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "time_created"),
					resource.TestCheckNoResourceAttr(resourceName, "version_date"),

					func(s *terraform.State) (err error) {
						resId, err = fromInstanceState(s, resourceName, "id")
						return err
					},
				),
			},

			// verify updates to updatable parameters
			{
				Config: config + `
variable "policy_defined_tags_value" { default = "updatedValue" }
variable "policy_description" { default = "description2" }
variable "policy_freeform_tags" { default = {"Department"= "Accounting"} }
variable "policy_name" { default = "LaunchInstances" }
variable "policy_version_date" { default = "2018-01-01" }

                ` + compartmentIdVariableStr + PolicyResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "compartment_id", tenancyId),
					resource.TestCheckResourceAttr(resourceName, "defined_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "LaunchInstances"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "statements.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "time_created"),
					resource.TestCheckResourceAttr(resourceName, "version_date", "2018-01-01"),

					func(s *terraform.State) (err error) {
						resId2, err = fromInstanceState(s, resourceName, "id")
						if resId != resId2 {
							return fmt.Errorf("Resource recreated when it was supposed to be updated.")
						}
						return err
					},
				),
			},
			// verify datasource
			{
				Config: config + `
variable "policy_defined_tags_value" { default = "updatedValue" }
variable "policy_description" { default = "description2" }
variable "policy_freeform_tags" { default = {"Department"= "Accounting"} }
variable "policy_name" { default = "LaunchInstances" }
variable "policy_version_date" { default = "2018-01-01" }

data "oci_identity_policies" "test_policies" {
	#Required
	compartment_id = "${var.tenancy_ocid}"

    filter {
    	name = "id"
    	values = ["${oci_identity_policy.test_policy.id}"]
    }
}
                ` + compartmentIdVariableStr + PolicyResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "compartment_id", tenancyId),

					resource.TestCheckResourceAttr(datasourceName, "policies.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "policies.0.compartment_id", tenancyId),
					resource.TestCheckResourceAttr(datasourceName, "policies.0.defined_tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "policies.0.description", "description2"),
					resource.TestCheckResourceAttr(datasourceName, "policies.0.freeform_tags.%", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "policies.0.id"),
					resource.TestCheckResourceAttr(datasourceName, "policies.0.name", "LaunchInstances"),
					resource.TestCheckResourceAttrSet(datasourceName, "policies.0.state"),
					resource.TestCheckResourceAttr(datasourceName, "policies.0.statements.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "policies.0.time_created"),
					resource.TestCheckResourceAttr(datasourceName, "policies.0.version_date", "2018-01-01"),
				),
			},
			// verify resource import
			{
				Config:            config,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// ETag, lastUpdateETag, and policyHash are non-API fields that
					// get computed during resource Create/Update but omitted from Get calls.
					// These are internally used for diff suppression and not needed for imports.
					// Omit them in the import verification.
					"ETag",
					"lastUpdateETag",
					"policyHash",
				},
				ResourceName: resourceName,
			},
		},
	})
}

func testAccCheckIdentityPolicyDestroy(s *terraform.State) error {
	noResourceFound := true
	client := testAccProvider.Meta().(*OracleClients).identityClient
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "oci_identity_policy" {
			noResourceFound = false
			request := oci_identity.GetPolicyRequest{}

			tmp := rs.Primary.ID
			request.PolicyId = &tmp

			response, err := client.GetPolicy(context.Background(), request)

			if err == nil {
				deletedLifecycleStates := map[string]bool{
					string(oci_identity.PolicyLifecycleStateDeleted): true,
				}
				if _, ok := deletedLifecycleStates[string(response.LifecycleState)]; !ok {
					//resource lifecycle state is not in expected deleted lifecycle states.
					return fmt.Errorf("resource lifecycle state: %s is not in expected deleted lifecycle states", response.LifecycleState)
				}
				//resource lifecycle state is in expected deleted lifecycle states. continue with next one.
				continue
			}

			//Verify that exception is for '404 not found'.
			if failure, isServiceError := common.IsServiceError(err); !isServiceError || failure.GetHTTPStatusCode() != 404 {
				return err
			}
		}
	}
	if noResourceFound {
		return fmt.Errorf("at least one resource was expected from the state file, but could not be found")
	}

	return nil
}
