package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	affiliatecontroller "mtg-price-checker-sg/controller/affiliatelinks"
	"mtg-price-checker-sg/store/affiliatelinks"

	"github.com/aws/aws-lambda-go/events"
)

const (
	affiliateLinksPublicPath = "/affiliate-links"
	affiliateLinksAdminPath  = "/admin/affiliate-links"
	affiliateAdminCORSMethods = "GET, POST, PUT, DELETE, OPTIONS"
)

type affiliateLinksResponse struct {
	Data []affiliatelinks.Link `json:"data"`
}

var newAffiliateService = func(ctx context.Context) (*affiliatecontroller.Service, error) {
	store, err := affiliatelinks.NewDynamoDBStore(ctx)
	if err != nil {
		return nil, err
	}

	uploader, err := affiliatelinks.NewImageUploader(ctx)
	if err != nil {
		return nil, err
	}

	return affiliatecontroller.NewService(store, uploader), nil
}

func AffiliateLinks(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	origin := request.Headers["origin"]

	if request.HTTPMethod == "OPTIONS" {
		return optionsResponseWithMethods(origin, "GET, OPTIONS")
	}

	if request.HTTPMethod != http.MethodGet {
		return errorResponse(apiRes, origin, "method not allowed", http.StatusMethodNotAllowed)
	}

	service, err := newAffiliateService(ctx)
	if err != nil {
		return errorResponse(apiRes, origin, "affiliate links are not configured", http.StatusServiceUnavailable)
	}

	links, err := service.ListActive(ctx)
	if err != nil {
		return errorResponse(apiRes, origin, "failed to load affiliate links", http.StatusInternalServerError)
	}
	if links == nil {
		links = []affiliatelinks.Link{}
	}

	return jsonResponse(apiRes, origin, http.StatusOK, affiliateLinksResponse{Data: links})
}

func AdminAffiliateLinks(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	origin := request.Headers["origin"]
	path := normalizeAPIPath(request.Path)

	if request.HTTPMethod == "OPTIONS" {
		return optionsResponseWithMethods(origin, affiliateAdminCORSMethods)
	}

	if !isAffiliateAdminAuthorized(request.Headers) {
		return errorResponseWithMethods(apiRes, origin, "unauthorized", http.StatusUnauthorized, affiliateAdminCORSMethods)
	}

	service, err := newAffiliateService(ctx)
	if err != nil {
		return errorResponseWithMethods(apiRes, origin, "affiliate links are not configured", http.StatusServiceUnavailable, affiliateAdminCORSMethods)
	}

	switch request.HTTPMethod {
	case http.MethodGet:
		links, listErr := service.ListAll(ctx)
		if listErr != nil {
			return errorResponseWithMethods(apiRes, origin, "failed to load affiliate links", http.StatusInternalServerError, affiliateAdminCORSMethods)
		}
		if links == nil {
			links = []affiliatelinks.Link{}
		}
		return jsonResponseWithMethods(apiRes, origin, http.StatusOK, affiliateLinksResponse{Data: links}, affiliateAdminCORSMethods)

	case http.MethodPost:
		var input affiliatecontroller.CreateInput
		if err := json.Unmarshal([]byte(request.Body), &input); err != nil {
			return errorResponseWithMethods(apiRes, origin, "invalid request body", http.StatusBadRequest, affiliateAdminCORSMethods)
		}
		link, createErr := service.Create(ctx, input)
		if createErr != nil {
			return errorResponseWithMethods(apiRes, origin, createErr.Error(), http.StatusBadRequest, affiliateAdminCORSMethods)
		}
		return jsonResponseWithMethods(apiRes, origin, http.StatusCreated, link, affiliateAdminCORSMethods)

	case http.MethodPut:
		id := affiliateLinkIDFromPath(path)
		if id == "" {
			return errorResponseWithMethods(apiRes, origin, "link id is required", http.StatusBadRequest, affiliateAdminCORSMethods)
		}
		var input affiliatecontroller.UpdateInput
		if err := json.Unmarshal([]byte(request.Body), &input); err != nil {
			return errorResponseWithMethods(apiRes, origin, "invalid request body", http.StatusBadRequest, affiliateAdminCORSMethods)
		}
		link, updateErr := service.Update(ctx, id, input)
		if updateErr != nil {
			if updateErr.Error() == "affiliate link not found" {
				return errorResponseWithMethods(apiRes, origin, updateErr.Error(), http.StatusNotFound, affiliateAdminCORSMethods)
			}
			return errorResponseWithMethods(apiRes, origin, updateErr.Error(), http.StatusBadRequest, affiliateAdminCORSMethods)
		}
		return jsonResponseWithMethods(apiRes, origin, http.StatusOK, link, affiliateAdminCORSMethods)

	case http.MethodDelete:
		id := affiliateLinkIDFromPath(path)
		if id == "" {
			return errorResponseWithMethods(apiRes, origin, "link id is required", http.StatusBadRequest, affiliateAdminCORSMethods)
		}
		if deleteErr := service.Delete(ctx, id); deleteErr != nil {
			if deleteErr.Error() == "affiliate link not found" {
				return errorResponseWithMethods(apiRes, origin, deleteErr.Error(), http.StatusNotFound, affiliateAdminCORSMethods)
			}
			return errorResponseWithMethods(apiRes, origin, "failed to delete affiliate link", http.StatusInternalServerError, affiliateAdminCORSMethods)
		}
		return jsonResponseWithMethods(apiRes, origin, http.StatusNoContent, map[string]string{}, affiliateAdminCORSMethods)

	default:
		return errorResponseWithMethods(apiRes, origin, "method not allowed", http.StatusMethodNotAllowed, affiliateAdminCORSMethods)
	}
}

func routeAffiliateRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := normalizeAPIPath(request.Path)
	switch {
	case path == affiliateLinksPublicPath:
		return AffiliateLinks(ctx, request)
	case path == affiliateLinksAdminPath || strings.HasPrefix(path, affiliateLinksAdminPath+"/"):
		return AdminAffiliateLinks(ctx, request)
	default:
		var apiRes events.APIGatewayProxyResponse
		return errorResponse(apiRes, request.Headers["origin"], "not found", http.StatusNotFound)
	}
}

func normalizeAPIPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if idx := strings.Index(path, "/admin/"); idx >= 0 {
		return strings.TrimSuffix(path[idx:], "/")
	}
	if idx := strings.Index(path, "/affiliate-links"); idx >= 0 {
		return strings.TrimSuffix(path[idx:], "/")
	}
	return strings.TrimSuffix(path, "/")
}

func affiliateLinkIDFromPath(path string) string {
	path = strings.TrimPrefix(path, affiliateLinksAdminPath)
	path = strings.TrimPrefix(path, "/")
	return strings.TrimSpace(path)
}
