package coinank_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"nofx/provider/coinank"
	"nofx/provider/coinank/coinank_enum"
	"strconv"
	"time"
)

const MainApiUrl = "https://api.coinank.com"

// CoinAnkError custom error type with more details
type CoinAnkError struct {
	Symbol     string
	Code       string
	HTTPStatus int
	Response   string
	IsNotFound bool // true if symbol is not supported
}

func (e *CoinAnkError) Error() string {
	if e.IsNotFound {
		return fmt.Sprintf("CoinAnk: symbol %s not supported (code=%s)", e.Symbol, e.Code)
	}
	return fmt.Sprintf("CoinAnk API error for %s (code=%s, status=%d)", e.Symbol, e.Code, e.HTTPStatus)
}

// Kline open free kline from coinank
func Kline(ctx context.Context, symbol string, exchange coinank_enum.Exchange, ts int64, side coinank_enum.Side, size int,
	interval coinank_enum.Interval) ([]coinank.KlineResult, error) {
	paramsMap := make(map[string]string, 6)
	paramsMap["symbol"] = symbol
	paramsMap["exchange"] = string(exchange)
	paramsMap["side"] = string(side)
	paramsMap["size"] = strconv.Itoa(size)
	paramsMap["ts"] = strconv.FormatInt(ts, 10)
	paramsMap["interval"] = string(interval)
	resp, httpStatus, err := get(ctx, "/api/kline/list/open", paramsMap)
	if err != nil {
		return nil, err
	}
	var result coinank.CoinankResponse[[][]float64]
	err = json.Unmarshal([]byte(resp), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (response: %s)", err, resp)
	}
	if !result.Success {
		// Check if it's a "not found" error (symbol not supported)
		// Common error codes for unsupported symbols: "0", "404", "SYMBOL_NOT_FOUND", etc.
		isNotFound := result.Code == "0" || result.Code == "404" ||
			result.Code == "SYMBOL_NOT_FOUND" || result.Data == nil

		return nil, &CoinAnkError{
			Symbol:     symbol,
			Code:       result.Code,
			HTTPStatus: httpStatus,
			Response:   resp,
			IsNotFound: isNotFound,
		}
	}
	klines := make([]coinank.KlineResult, len(result.Data))
	for i, k := range result.Data {
		klines[i].StartTime = int64(k[0] + 0.001)
		klines[i].EndTime = int64(k[1] + 0.001)
		klines[i].Open = k[2]
		klines[i].Close = k[3]
		klines[i].High = k[4]
		klines[i].Low = k[5]
		klines[i].Volume = k[6]
		klines[i].Quantity = k[7]
		klines[i].Count = k[8]
	}
	return klines, nil
}

func get(ctx context.Context, path string, paramsMap map[string]string) (string, int, error) {
	data := url.Values{}
	for key, value := range paramsMap {
		data.Add(key, value)
	}
	fullURL := fmt.Sprintf("%s%s?%s", MainApiUrl, path, data.Encode())
	request, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.Do(request)
	if err != nil {
		return "", 0, fmt.Errorf("HTTP request failed (url: %s): %w", fullURL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}
	return string(body), resp.StatusCode, nil
}

var client = &http.Client{
	Timeout: 30 * time.Second,
}
