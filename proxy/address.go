package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ekomobile/dadata/v2/api/suggest"
	"github.com/ekomobile/dadata/v2/client"
	"net/http"
	"net/url"
	"strings"
)

type GeoService struct {
	api       *suggest.Api
	apiKey    string
	secretKey string
}

type GeoProvider interface {
	AddressSearch(input string) ([]*Address, error)
	GeoCode(lat, lng string) ([]*Address, error)
}

func NewGeoService(apiKey, secretKey string) *GeoService {
	var err error
	endpointUrl, err := url.Parse("https://suggestions.dadata.ru/suggestions/api/4_1/rs/")
	if err != nil {
		return nil
	}

	creds := client.Credentials{
		ApiKeyValue:    apiKey,
		SecretKeyValue: secretKey,
	}

	api := suggest.Api{
		Client: client.NewClient(endpointUrl, client.WithCredentialProvider(&creds)),
	}

	return &GeoService{
		api:       &api,
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

type Address struct {
	City   string `json:"city"`
	Street string `json:"street"`
	House  string `json:"house"`
	Lat    string `json:"lat"`
	Lon    string `json:"lon"`
}

func (g *GeoService) AddressSearch(input string) ([]*Address, error) {
	var res []*Address
	rawRes, err := g.api.Address(context.Background(), &suggest.RequestParams{Query: input})
	if err != nil {
		return nil, err
	}

	for _, r := range rawRes {
		if r.Data.City == "" || r.Data.Street == "" {
			continue
		}
		res = append(res, &Address{City: r.Data.City, Street: r.Data.Street, House: r.Data.House, Lat: r.Data.GeoLat, Lon: r.Data.GeoLon})
	}

	return res, nil
}

func (g *GeoService) GeoCode(lat, lng string) ([]*Address, error) {
	httpClient := &http.Client{}
	var data = strings.NewReader(fmt.Sprintf(`{"lat": %s, "lon": %s}`, lat, lng))
	req, err := http.NewRequest("POST", "https://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address", data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", g.apiKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	var geoCode GeoCode

	err = json.NewDecoder(resp.Body).Decode(&geoCode)
	if err != nil {
		return nil, err
	}
	var res []*Address
	for _, r := range geoCode.Suggestions {
		var address Address
		address.City = string(r.Data.City)
		address.Street = string(r.Data.Street)
		address.House = r.Data.House
		address.Lat = r.Data.GeoLat
		address.Lon = r.Data.GeoLon

		res = append(res, &address)
	}

	return res, nil
}

type GeocodeRequest struct {
	Lat string `json:"lat" example:"55.878"`
	Lng string `json:"lng" example:"37.653"`
}

type GeocodeResponse struct {
	Addresses []*Address `json:"addresses"`
}

// @Summary receive Address by GeoCode
// @Tags GeoCode
// @Description Request structure for geocoding addresses
// @ID geo
// @Accept json
// @Produce json
// @Param input body GeocodeRequest true "Handle Address by GeoCode"
// @Success 200 {object} GeocodeResponse
// @failure 400 {string} string "Empty Query"
// @failure 500 {string} string "Internal Server Error"
// @Router /api/address/geocode [post]
func (g *GeoService) handleAddressGeocode(w http.ResponseWriter, r *http.Request) {
	var geoReq GeocodeRequest
	if err := json.NewDecoder(r.Body).Decode(&geoReq); err != nil || geoReq.Lng == "" || geoReq.Lat == "" {
		http.Error(w, "Empty Query", http.StatusBadRequest)
		return
	}

	geo, err := g.GeoCode(geoReq.Lat, geoReq.Lng)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := GeocodeResponse{geo}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type SearchRequest struct {
	Query string `json:"query" example:"мск сухонска 11/-89"`
}
type SearchResponse struct {
	Addresses []*Address `json:"addresses"`
}

// @Summary receive Information by Address
// @Tags AddressSearch
// @Description Receive Information by Address
// @ID addSearch
// @Accept json
// @Produce json
// @Param input body SearchRequest true "Receive information by Address"
// @Success 200 {object} SearchResponse
// @failure 400 {string} string "Empty Query"
// @failure 500 {string} string "Internal Server Error"
// @Router /address/search [post]
func (g *GeoService) handleAddressSearch(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Query == "" {
		http.Error(w, "Empty Query", http.StatusBadRequest)
		return
	}

	addresses, err := g.AddressSearch(req.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := SearchResponse{addresses}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

}
