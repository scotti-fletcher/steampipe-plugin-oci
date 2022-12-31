package oci

import (
		"context"
		"github.com/oracle/oci-go-sdk/v65/common"
		"github.com/oracle/oci-go-sdk/v65/artifacts"
		"github.com/turbot/go-kit/types"
		"github.com/turbot/steampipe-plugin-sdk/v4/grpc/proto"
		"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
		"github.com/turbot/steampipe-plugin-sdk/v4/plugin/transform"
		"strings"
)

//// TABLE DEFINITION
func tableArtifactsContainerRepository(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "oci_artifacts_container_repository",
		Description:      "OCI Container Repository",
		DefaultTransform: transform.FromCamel(),
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate: getArtifactsContainerRepository,
		},
		List: &plugin.ListConfig{
			Hydrate: listArtifactsContainerRepositories,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:		"compartment_id",
					Require: plugin.Optional,
				},
				{
					Name:		"display_name",
					Require: plugin.Optional,
				},
				{
					Name:		"is_public",
					Require: plugin.Optional,
				},
				{
					Name: 	"lifecycle_state",
					Require: plugin.Optional,
				},
			},
		},
		GetMatrixItemFunc: BuildCompartementRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "created_by",
				Description: "The id of the user or principal that created the resource.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getArtifactsContainerRepository,
			},
			{
				Name:        "display_name",
				Description: "The container repository name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the container repository.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "image_count",
				Description: "Total number of images.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "is_immutable",
				Description: "Whether the repository is immutable. Images cannot be overwritten in an immutable repository.",
				Type:        proto.ColumnType_BOOL,
				Hydrate:     getArtifactsContainerRepository,
			},
			{
				Name:        "is_public",
				Description: "Whether the repository is public. A public repository allows unauthenticated access.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "layer_count",
				Description: "Total number of layers.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "layers_size_in_bytes",
				Description: "Total storage in bytes consumed by layers.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "lifecycle_state",
				Description: "The current state of the container repository.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "billable_size_in_g_bs",
				Description: "Total storage size in GBs that will be charged.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "time_last_pushed",
				Description: "An RFC 3339 timestamp indicating when an image was last pushed to the repository.",
				Type:        proto.ColumnType_TIMESTAMP,
				Hydrate:     getArtifactsContainerRepository,
				Transform:	transform.FromField("TimeLastPushed.Time"),
			},
			{
				Name:        "time_created",
				Description: "Time that Container Repository was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("TimeCreated.Time"),
			},

  		// Standard Steampipe columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,

				Transform:   transform.FromField("DisplayName"),
			},

			// Standard OCI columns
			{
				Name:        "compartment_id",
				Description: ColumnDescriptionCompartment,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CompartmentId"),
			},
			{
				Name:        "tenant_id",
				Description: ColumnDescriptionTenant,
				Type:        proto.ColumnType_STRING,
				Hydrate:     plugin.HydrateFunc(getTenantId).WithCache(),
				Transform:   transform.FromValue(),
			},
		},
	}
}

//// LIST FUNCTION
func listArtifactsContainerRepositories(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	region := plugin.GetMatrixItem(ctx)[matrixKeyRegion].(string)
	compartment := plugin.GetMatrixItem(ctx)[matrixKeyCompartment].(string)
	logger.Debug("listArtifactsContainerRepositories", "Compartment", compartment, "OCI_REGION", region)


	equalQuals := d.KeyColumnQuals
	// Return nil, if given compartment_id doesn't match
	if equalQuals["compartment_id"] != nil && compartment != equalQuals["compartment_id"].GetStringValue() {
		return nil, nil
	}
	// Create Session
	session, err := artifactsService(ctx, d, region)
	if err != nil {
		return nil, err
	}

	//Build request parameters
	request := buildArtifactsContainerRepositoryFilters(equalQuals)
	request.CompartmentId = types.String(compartment)
	request.Limit = types.Int(100)
	request.RequestMetadata = common.RequestMetadata{
		RetryPolicy: getDefaultRetryPolicy(d.Connection),
	}

	limit := d.QueryContext.Limit
	if d.QueryContext.Limit != nil {
		if *limit < int64(*request.Limit) {
			request.Limit = types.Int(int(*limit))
		}
	}

	pagesLeft := true
	for pagesLeft {
		response, err := session.ArtifactsClient.ListContainerRepositories(ctx, request)
		if err != nil {
			return nil, err
		}
		for _, respItem := range response.Items {
			d.StreamListItem(ctx, respItem)

			// Context can be cancelled due to manual cancellation or the limit has been hit
			if d.QueryStatus.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
		if response.OpcNextPage != nil {
			request.Page = response.OpcNextPage
		} else {
			pagesLeft = false
		}
	}

	return nil, err
}

//// HYDRATE FUNCTION
func getArtifactsContainerRepository(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	region := plugin.GetMatrixItem(ctx)[matrixKeyRegion].(string)
	compartment := plugin.GetMatrixItem(ctx)[matrixKeyCompartment].(string)
	logger.Debug("getArtifactsContainerRepository", "Compartment", compartment, "OCI_REGION", region)



	var id string
	if h.Item != nil {
		id = *h.Item.(artifacts.ContainerRepositorySummary).Id
	} else {
		id = d.KeyColumnQuals["id"].GetStringValue()
		if !strings.HasPrefix(compartment, "ocid1.tenancy.oc1") {
			return nil, nil
		}
	}

	// handle empty id in get call
	if id == "" {
		return nil, nil
	}

	// Create Session

	session, err := artifactsService(ctx, d, region)
	if err != nil {
		logger.Error("getArtifactsContainerRepository", "error_ArtifactsService", err)
		return nil, err
	}

	request := artifacts.GetContainerRepositoryRequest{
		RepositoryId: types.String(id),
		RequestMetadata: common.RequestMetadata{
			RetryPolicy: getDefaultRetryPolicy(d.Connection),
		},
	}

	response, err := session.ArtifactsClient.GetContainerRepository(ctx, request)
	if err != nil {
		return nil, err
	}
	return response.ContainerRepository, nil
}


// Build additional filters
func buildArtifactsContainerRepositoryFilters(equalQuals plugin.KeyColumnEqualsQualMap) artifacts.ListContainerRepositoriesRequest {
	request := artifacts.ListContainerRepositoriesRequest{}

		if equalQuals["compartment_id"] != nil {
		request.CompartmentId = types.String(equalQuals["compartment_id"].GetStringValue())
		}

		if equalQuals["display_name"] != nil {
		request.DisplayName = types.String(equalQuals["display_name"].GetStringValue())
		}
		if equalQuals["is_public"] != nil {
			request.IsPublic = types.Bool(equalQuals["is_public"].GetBoolValue())
		}
		if equalQuals["lifecycle_state"] != nil {
			request.LifecycleState = types.String(equalQuals["lifecycle_state"].GetStringValue())
		}

	return request
}
