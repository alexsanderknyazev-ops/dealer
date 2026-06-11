package main

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/dealer/dealer/auth-service/internal/config"
	"github.com/dealer/dealer/auth-service/internal/routepaths"
)

func registerDomainAPIProxy(mux *http.ServeMux, cfg *config.Config, logger *slog.Logger) {
	registerTelemetryProxy(mux, cfg, logger)

	if cfg.GatewayServiceURL != "" {
		targetURL, err := url.Parse(cfg.GatewayServiceURL)
		if err != nil {
			logger.Error("invalid GATEWAY_SERVICE_URL", "url", cfg.GatewayServiceURL, "err", err)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		for _, prefix := range routepaths.GatewayProxyPrefixes() {
			mux.Handle(prefix, proxy)
			mux.Handle(prefix+"/", proxy)
		}
		logger.Info("proxying domain API via grpc-gateway", "target", cfg.GatewayServiceURL)
		return
	}

	type route struct {
		baseURL string
		paths   []string
		label   string
	}
	legacy := []route{
		{cfg.CustomersServiceURL, []string{routepaths.APICustomers, routepaths.APICustomersPrefix}, "customers"},
		{cfg.VehiclesServiceURL, []string{routepaths.APIVehicles, routepaths.APIVehiclesPrefix}, "vehicles"},
		{cfg.DealsServiceURL, []string{routepaths.APIDeals, routepaths.APIDealsPrefix}, "deals"},
		{cfg.PartsServiceURL, []string{routepaths.APIParts, routepaths.APIPartsPrefix}, "parts"},
		{cfg.BrandsServiceURL, []string{routepaths.APIBrands, routepaths.APIBrandsPrefix}, "brands"},
		{
			cfg.DealerPointsServiceURL,
			[]string{
				routepaths.APIDealerPoints, routepaths.APIDealerPointsPre,
				routepaths.APILegalEntities, routepaths.APILegalEntitiesPre,
				routepaths.APIWarehouses, routepaths.APIWarehousesPrefix,
			},
			"dealer-points",
		},
	}
	for _, r := range legacy {
		if strings.TrimSpace(r.baseURL) == "" {
			continue
		}
		targetURL, err := url.Parse(r.baseURL)
		if err != nil {
			logger.Error("invalid service proxy URL", "service", r.label, "url", r.baseURL, "err", err)
			continue
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		for _, p := range r.paths {
			mux.Handle(p, proxy)
		}
		logger.Info("proxying API", "service", r.label, "target", r.baseURL)
	}
}

func registerTelemetryProxy(mux *http.ServeMux, cfg *config.Config, logger *slog.Logger) {
	if strings.TrimSpace(cfg.ErrorsIngestServiceURL) == "" {
		return
	}
	targetURL, err := url.Parse(cfg.ErrorsIngestServiceURL)
	if err != nil {
		logger.Error("invalid ERRORS_INGEST_SERVICE_URL", "url", cfg.ErrorsIngestServiceURL, "err", err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	mux.Handle(routepaths.APITelemetry, proxy)
	mux.Handle(routepaths.APITelemetryPrefix, proxy)
	logger.Info("proxying telemetry to errors-ingest", "target", cfg.ErrorsIngestServiceURL)
}
