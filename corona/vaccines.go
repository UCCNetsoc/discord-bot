package corona

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Vaccines from the HSE arcGIS.
type Vaccines struct {
	First int
	Date  time.Time
}

// GetVaccines will query HSE arcGIS and return vaccine stats.
func GetVaccines() (*Vaccines, error) {
	resp, err := http.Get(covidVaccines)
	if err != nil {
		return nil, err
	}
	vaccines := &struct {
		Features []struct {
			Attributes struct {
				First int `json:"total_number_of_1st_dose_admini"`
				Date  int `json:"data_relevent_up_to_date"`
			} `json:"attributes"`
		} `json:"features"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(vaccines); err != nil {
		return nil, err
	}
	if len(vaccines.Features) < 1 {
		return nil, errors.New("no features")
	}
	return &Vaccines{
		First: vaccines.Features[0].Attributes.First,
		Date:  time.Unix(int64(vaccines.Features[0].Attributes.Date)/1000, 0),
	}, nil
}
