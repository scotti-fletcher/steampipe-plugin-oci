package oci

import (
	"context"
	"github.com/oracle/oci-go-sdk/v65/certificatesmanagement"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v4/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin/transform"
	"strings"
)

// TABLE DEFINITION
func tableCertificatesManagementAssociation(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:             "oci_certificates_management_association",
		Description:      "OCI Association",
		DefaultTransform: transform.FromCamel(),
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getCertificatesManagementAssociation,
		},
		List: &plugin.ListConfig{
			Hydrate: listCertificatesManagementAssociations,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "compartment_id",
					Require: plugin.Optional,
				},
				{
					Name:    "certificates_resource_id",
					Require: plugin.Optional,
				},
				{
					Name:    "associated_resource_id",
					Require: plugin.Optional,
				},
				{
					Name:    "name",
					Require: plugin.Optional,
				},
				{
					Name:    "association_type",
					Require: plugin.Optional,
				},
			},
		},
		GetMatrixItemFunc: BuildCompartementRegionList,
		Columns: []*plugin.Column{
			{
				Name:        "id",
				Description: "The OCID of the association.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "A user-friendly name generated by the service for the association, expressed in a format that follows the pattern: [certificatesResourceEntityType]-[associatedResourceEntityType]-UUID.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "lifecycle_state",
				Description: "The current lifecycle state of the association.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "certificates_resource_id",
				Description: "The OCID of the certificate-related resource associated with another Oracle Cloud Infrastructure resource.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "associated_resource_id",
				Description: "The OCID of the associated resource.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "association_type",
				Description: "Type of the association.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "time_created",
				Description: "Time that the Association was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("TimeCreated.Time"),
			},

			// Standard Steampipe columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
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

// LIST FUNCTION
func listCertificatesManagementAssociations(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	region := plugin.GetMatrixItem(ctx)[matrixKeyRegion].(string)
	compartment := plugin.GetMatrixItem(ctx)[matrixKeyCompartment].(string)
	logger.Debug("listCertificatesManagementAssociations", "Compartment", compartment, "OCI_REGION", region)

	equalQuals := d.KeyColumnQuals
	// Return nil, if given compartment_id doesn't match
	if equalQuals["compartment_id"] != nil && compartment != equalQuals["compartment_id"].GetStringValue() {
		return nil, nil
	}
	// Create Session
	session, err := certificatesManagementService(ctx, d, region)
	if err != nil {
		return nil, err
	}

	//Build request parameters
	request := buildListCertificatesManagementAssociationFilters(equalQuals)
	request.CompartmentId = types.String(compartment)
	request.Limit = types.Int(20)
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
		response, err := session.CertificatesManagementClient.ListAssociations(ctx, request)
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

// HYDRATE FUNCTION
func getCertificatesManagementAssociation(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	region := plugin.GetMatrixItem(ctx)[matrixKeyRegion].(string)
	compartment := plugin.GetMatrixItem(ctx)[matrixKeyCompartment].(string)
	logger.Debug("getCertificatesManagementAssociation", "Compartment", compartment, "OCI_REGION", region)
	if !strings.HasPrefix(compartment, "ocid1.tenancy.oc1") {
		return nil, nil
	}

	request := buildGetCertificatesManagementAssociationFilters(d.KeyColumnQuals, h)

	// Create Session
	session, err := certificatesManagementService(ctx, d, region)
	if err != nil {
		logger.Error("getCertificatesManagementAssociation", "error_CertificatesManagementService", err)
		return nil, err
	}
	request.RequestMetadata = common.RequestMetadata{
		RetryPolicy: getDefaultRetryPolicy(d.Connection),
	}

	response, err := session.CertificatesManagementClient.GetAssociation(ctx, request)
	if err != nil {
		return nil, err
	}
	return response.Association, nil
}

// Build additional list filters
func buildListCertificatesManagementAssociationFilters(equalQuals plugin.KeyColumnEqualsQualMap) certificatesmanagement.ListAssociationsRequest {
	request := certificatesmanagement.ListAssociationsRequest{}

	if equalQuals["compartment_id"] != nil {
		request.CompartmentId = types.String(equalQuals["compartment_id"].GetStringValue())
	}

	if equalQuals["certificates_resource_id"] != nil {
		request.CertificatesResourceId = types.String(equalQuals["certificates_resource_id"].GetStringValue())
	}

	if equalQuals["associated_resource_id"] != nil {
		request.AssociatedResourceId = types.String(equalQuals["associated_resource_id"].GetStringValue())
	}

	if equalQuals["name"] != nil {
		request.Name = types.String(equalQuals["name"].GetStringValue())
	}

	if equalQuals["association_type"] != nil {
		request.AssociationType = certificatesmanagement.ListAssociationsAssociationTypeEnum(equalQuals["association_type"].GetStringValue())
	}

	return request
}

// Build additional filters
func buildGetCertificatesManagementAssociationFilters(equalQuals plugin.KeyColumnEqualsQualMap, h *plugin.HydrateData) certificatesmanagement.GetAssociationRequest {
	request := certificatesmanagement.GetAssociationRequest{}

	if h.Item != nil {
		request.AssociationId = h.Item.(certificatesmanagement.AssociationSummary).Id
	} else {
		request.AssociationId = types.String(equalQuals["id"].GetStringValue())
	}

	return request
}
