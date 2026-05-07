// Package handler contains the REST handlers for the API gateway.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Gvinay90/ad-bidding-platform/api-gateway/client"
	analyticspb "github.com/Gvinay90/ad-bidding-platform/proto/analytics"
	bidderpb "github.com/Gvinay90/ad-bidding-platform/proto/bidder"
	campaignpb "github.com/Gvinay90/ad-bidding-platform/proto/campaign"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const rpcTimeout = 200 * time.Millisecond

// Handler holds gRPC clients and serves HTTP requests.
type Handler struct {
	clients *client.Clients
}

// New creates a Handler.
func New(c *client.Clients) *Handler { return &Handler{clients: c} }

// ── Swagger response models ────────────────────────────────────────────────

// Campaign is the JSON representation of a campaign.
type Campaign struct {
	ID            string `json:"id"`
	AdvertiserID  string `json:"advertiser_id"`
	Name          string `json:"name"`
	BudgetCents   int64  `json:"budget_cents"`
	BidPriceCents int64  `json:"bid_price_cents"`
	Geo           string `json:"geo"`
	Device        string `json:"device"`
	Category      string `json:"category"`
	Status        string `json:"status"`
}

// BidResponse is the JSON representation of a bid result.
type BidResponse struct {
	HasBid            bool   `json:"has_bid"`
	WinningCampaignID string `json:"winning_campaign_id"`
	PriceCents        int64  `json:"price_cents"`
}

// StatsResponse is the JSON representation of campaign analytics.
type StatsResponse struct {
	CampaignID  string `json:"campaign_id"`
	Wins        int64  `json:"wins"`
	Bids        int64  `json:"bids"`
	SpendCents  int64  `json:"spend_cents"`
}

// ErrorResponse wraps an error message.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ── Request models ─────────────────────────────────────────────────────────

// CreateCampaignRequest is the JSON body for POST /campaigns.
type CreateCampaignRequest struct {
	AdvertiserID  string `json:"advertiser_id" example:"adv-001"`
	Name          string `json:"name" example:"Summer Sale"`
	BudgetCents   int64  `json:"budget_cents" example:"100000"`
	BidPriceCents int64  `json:"bid_price_cents" example:"500"`
	Geo           string `json:"geo" example:"US"`
	Device        string `json:"device" example:"mobile"`
	Category      string `json:"category" example:"sports"`
}

// UpdateCampaignRequest is the JSON body for PUT /campaigns/{id}.
type UpdateCampaignRequest struct {
	Name          string `json:"name" example:"Updated Name"`
	BudgetCents   int64  `json:"budget_cents" example:"200000"`
	BidPriceCents int64  `json:"bid_price_cents" example:"750"`
	Geo           string `json:"geo" example:"US"`
	Device        string `json:"device" example:"desktop"`
	Category      string `json:"category" example:"tech"`
	Status        string `json:"status" example:"active"`
}

// BidRequest is the JSON body for POST /bid.
type BidRequest struct {
	Geo      string `json:"geo" example:"US"`
	Device   string `json:"device" example:"mobile"`
	Category string `json:"category" example:"sports"`
}

// MessageResponse wraps a plain message string.
type MessageResponse struct {
	Message string `json:"message"`
}

// HealthResponse is the response for the health endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// ── helpers ────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	st, _ := status.FromError(err)
	httpCode := grpcToHTTP(st.Code())
	writeJSON(w, httpCode, ErrorResponse{Error: st.Message()})
}

func grpcToHTTP(c codes.Code) int {
	switch c {
	case codes.NotFound:
		return http.StatusNotFound
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func rpcCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), rpcTimeout)
}

func pbCampaignToJSON(c *campaignpb.Campaign) Campaign {
	return Campaign{
		ID:            c.Id,
		AdvertiserID:  c.AdvertiserId,
		Name:          c.Name,
		BudgetCents:   c.BudgetCents,
		BidPriceCents: c.BidPriceCents,
		Geo:           c.Geo,
		Device:        c.Device,
		Category:      c.Category,
		Status:        c.Status,
	}
}

// ── Campaign handlers ──────────────────────────────────────────────────────

// Health godoc
//
//	@Summary		Health check
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/healthz [get]
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// CreateCampaign godoc
//
//	@Summary		Create a campaign
//	@Tags			campaigns
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateCampaignRequest	true	"Campaign payload"
//	@Success		201		{object}	Campaign
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/campaigns [post]
func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}
	ctx, cancel := rpcCtx(r)
	defer cancel()
	resp, err := h.clients.Campaign.CreateCampaign(ctx, &campaignpb.CreateCampaignRequest{
		AdvertiserId:  req.AdvertiserID,
		Name:          req.Name,
		BudgetCents:   req.BudgetCents,
		BidPriceCents: req.BidPriceCents,
		Geo:           req.Geo,
		Device:        req.Device,
		Category:      req.Category,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, pbCampaignToJSON(resp.Campaign))
}

// GetCampaign godoc
//
//	@Summary		Get a campaign by ID
//	@Tags			campaigns
//	@Produce		json
//	@Param			id	path		string	true	"Campaign ID"
//	@Success		200	{object}	Campaign
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/campaigns/{id} [get]
func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, cancel := rpcCtx(r)
	defer cancel()
	resp, err := h.clients.Campaign.GetCampaign(ctx, &campaignpb.GetCampaignRequest{Id: id})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pbCampaignToJSON(resp.Campaign))
}

// ListCampaigns godoc
//
//	@Summary		List campaigns for an advertiser
//	@Tags			campaigns
//	@Produce		json
//	@Param			advertiser_id	query		string	false	"Filter by advertiser ID"
//	@Success		200				{array}		Campaign
//	@Failure		500				{object}	ErrorResponse
//	@Router			/campaigns [get]
func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	advID := r.URL.Query().Get("advertiser_id")
	ctx, cancel := rpcCtx(r)
	defer cancel()
	resp, err := h.clients.Campaign.ListCampaigns(ctx, &campaignpb.ListCampaignsRequest{AdvertiserId: advID})
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]Campaign, 0, len(resp.Campaigns))
	for _, c := range resp.Campaigns {
		out = append(out, pbCampaignToJSON(c))
	}
	writeJSON(w, http.StatusOK, out)
}

// UpdateCampaign godoc
//
//	@Summary		Update a campaign
//	@Tags			campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Campaign ID"
//	@Param			body	body		UpdateCampaignRequest	true	"Fields to update"
//	@Success		200		{object}	Campaign
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/campaigns/{id} [put]
func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req UpdateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}
	ctx, cancel := rpcCtx(r)
	defer cancel()
	resp, err := h.clients.Campaign.UpdateCampaign(ctx, &campaignpb.UpdateCampaignRequest{
		Id:            id,
		Name:          req.Name,
		BudgetCents:   req.BudgetCents,
		BidPriceCents: req.BidPriceCents,
		Geo:           req.Geo,
		Device:        req.Device,
		Category:      req.Category,
		Status:        req.Status,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pbCampaignToJSON(resp.Campaign))
}

// DeleteCampaign godoc
//
//	@Summary		Delete a campaign
//	@Tags			campaigns
//	@Produce		json
//	@Param			id	path		string	true	"Campaign ID"
//	@Success		200	{object}	MessageResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/campaigns/{id} [delete]
func (h *Handler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, cancel := rpcCtx(r)
	defer cancel()
	resp, err := h.clients.Campaign.DeleteCampaign(ctx, &campaignpb.DeleteCampaignRequest{Id: id})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, MessageResponse{Message: resp.Message})
}

// EvaluateBid godoc
//
//	@Summary		Submit a bid request and get the winning campaign
//	@Tags			bidder
//	@Accept			json
//	@Produce		json
//	@Param			body	body		BidRequest		true	"Targeting attributes"
//	@Success		200		{object}	BidResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/bid [post]
func (h *Handler) EvaluateBid(w http.ResponseWriter, r *http.Request) {
	var req BidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}
	ctx, cancel := rpcCtx(r)
	defer cancel()
	resp, err := h.clients.Bidder.EvaluateBid(ctx, &bidderpb.BidRequest{
		RequestId: uuid.New().String(),
		Geo:       req.Geo,
		Device:    req.Device,
		Category:  req.Category,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, BidResponse{
		HasBid:            resp.HasBid,
		WinningCampaignID: resp.WinningCampaignId,
		PriceCents:        resp.PriceCents,
	})
}

// GetCampaignStats godoc
//
//	@Summary		Get analytics stats for a campaign
//	@Tags			analytics
//	@Produce		json
//	@Param			id	path		string	true	"Campaign ID"
//	@Success		200	{object}	StatsResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/stats/{id} [get]
func (h *Handler) GetCampaignStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, cancel := rpcCtx(r)
	defer cancel()
	resp, err := h.clients.Analytics.GetCampaignStats(ctx, &analyticspb.CampaignStatsRequest{CampaignId: id})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, StatsResponse{
		CampaignID: resp.CampaignId,
		Wins:       resp.Wins,
		Bids:       resp.Bids,
		SpendCents: resp.SpendCents,
	})
}
