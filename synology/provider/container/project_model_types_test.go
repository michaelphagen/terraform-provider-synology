package container

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/synology-community/terraform-provider-synology/synology/provider/container/models"
)

// The model types are used to build state values, so they have to match the
// schema exactly or setting state fails at runtime.
func TestProjectModelTypesMatchSchema(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var sch resource.SchemaResponse
	(&ProjectResource{}).Schema(ctx, resource.SchemaRequest{}, &sch)
	schemaTypes := sch.Schema.Type().(basetypes.ObjectType).AttrTypes

	for name, want := range (models.ProjectResourceModel{}).AttrType() {
		t.Run(name, func(t *testing.T) {
			have, ok := schemaTypes[name]
			if !ok {
				t.Fatalf("model has %q but the schema does not", name)
			}
			if !want.Equal(have) {
				t.Errorf("type mismatch for %q\nmodel:  %s\nschema: %s", name, want, have)
			}
		})
	}
}

// A service's ulimits, as Terraform hands them over, have to decode into
// models.Ulimit.
func TestUlimitsDecode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var sch resource.SchemaResponse
	(&ProjectResource{}).Schema(ctx, resource.SchemaRequest{}, &sch)
	services := sch.Schema.Type().(basetypes.ObjectType).AttrTypes["services"].(basetypes.MapType)
	ulimitsType := services.ElemType.(basetypes.ObjectType).AttrTypes["ulimits"].(basetypes.MapType)
	ulimitType := ulimitsType.ElemType.(basetypes.ObjectType)

	nofile := types.ObjectValueMust(ulimitType.AttrTypes, map[string]attr.Value{
		"name":  types.StringValue("nofile"),
		"value": types.Int64Null(),
		"soft":  types.Int64Value(1024),
		"hard":  types.Int64Value(4096),
	})
	ulimits := types.MapValueMust(ulimitType, map[string]attr.Value{"nofile": nofile})

	have := map[string]models.Ulimit{}
	if diags := ulimits.ElementsAs(ctx, &have, true); diags.HasError() {
		t.Fatalf("ulimits did not decode: %v", diags)
	}
	if have["nofile"].Soft.ValueInt64() != 1024 || have["nofile"].Hard.ValueInt64() != 4096 {
		t.Errorf("wrong ulimit\nhave: %+v\nwant: soft 1024, hard 4096", have["nofile"])
	}
}
