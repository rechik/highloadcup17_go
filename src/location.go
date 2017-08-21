package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

type Location struct {
	Id          uint32          `json:"id"`
	Place       string          `json:"place"`
	Country     string          `json:"country"`
	City        string          `json:"city"`
	Distance    uint32          `json:"distance"`
	Mutex       sync.RWMutex    `json:"-"`
	VisitIdsMap map[uint32]bool `json:"-"`
}

type LocationsRepo struct {
	Collection map[uint32]*Location
	Mutex      sync.RWMutex
}

func (entity *Location) Update(data *JsonData, lock bool) bool {
	if lock {
		entity.Mutex.Lock()
		defer entity.Mutex.Unlock()
	}
	denormolize_in_visits := false
	for key, value := range *data {
		if value == nil {
			return false
		}
		switch key {
		case "id":
			entity.Id = uint32(value.(float64))
		case "place":
			entity.Place = value.(string)
		case "country":
			entity.Country = value.(string)
			denormolize_in_visits = true
		case "city":
			entity.City = value.(string)
		case "distance":
			entity.Distance = uint32(value.(float64))
			denormolize_in_visits = true
		}
	}
	if denormolize_in_visits {
		visits := entity.Visits(nil, nil, nil, nil, nil)
		for _, visit := range visits {
			visit.LocationCountry = entity.Country
			visit.LocationDistance = entity.Distance
			visit.Mutex.RUnlock()
		}
	}
	return true
}

func (entity *Location) to_json(w io.Writer) {
	entity.Mutex.RLock()
	json.NewEncoder(w).Encode(entity)
	entity.Mutex.RUnlock()
}

func (entity *Location) VisitIds() []uint32 {
	ids := make([]uint32, len(entity.VisitIdsMap))

	i := 0
	for id := range entity.VisitIdsMap {
		ids[i] = id
		i++
	}
	return ids
}

func BirthDateToAge(BirthDate int32) uint32 {
	now := int32(time.Now().Unix())
	age_ts := int64(now - BirthDate)
	age := uint32(time.Unix(age_ts, 0).Year() - 1970)
	return age
}

func (entity *Location) Visits(fromDate *uint32, toDate *uint32, fromAge *uint32, toAge *uint32, gender *string) []*Visit {
	visits, _ := Visits.FindAll(entity.VisitIds())
	filteredVisits := make([]*Visit, 0, len(visits))
	for _, visit := range visits {
		visit.Mutex.RLock()
		if fromDate != nil && visit.VisitedAt < *fromDate {
			visit.Mutex.RUnlock()
			continue
		}
		if toDate != nil && visit.VisitedAt > *toDate {
			visit.Mutex.RUnlock()
			continue
		}
		if fromAge != nil || toAge != nil {
			age := BirthDateToAge(visit.UserBirthDate)
			if fromAge != nil && age < *fromAge {
				visit.Mutex.RUnlock()
				continue
			}
			if toAge != nil && age >= *toAge {
				visit.Mutex.RUnlock()
				continue
			}
		}
		if gender != nil && visit.UserGender != *gender {
			visit.Mutex.RUnlock()
			continue
		}
		filteredVisits = append(filteredVisits, visit)
	}
	return filteredVisits
}

func (entity *Location) WriteAvgsJson(w io.Writer, fromDate *uint32, toDate *uint32, fromAge *uint32, toAge *uint32, gender *string) {
	entity.Mutex.RLock()

	filteredVisits := entity.Visits(fromDate, toDate, fromAge, toAge, gender)
	if len(filteredVisits) == 0 {
		w.Write([]byte("{\"avg\": 0}"))
	} else {
		marks_count := 0
		marks_sum := uint32(0)

		for _, visit := range filteredVisits {
			marks_sum += visit.Mark
			marks_count += 1
			visit.Mutex.RUnlock()
		}

		avg := float64(marks_sum) / float64(marks_count)
		avg_str := fmt.Sprintf("%.5f", avg)

		w.Write([]byte("{\"avg\": "))
		w.Write([]byte(avg_str))
		w.Write([]byte("}"))
	}
	entity.Mutex.RUnlock()
}

func NewLocationsRepo() LocationsRepo {
	return LocationsRepo{
		Collection: make(map[uint32]*Location),
		Mutex:      sync.RWMutex{}}
}

func (repo *LocationsRepo) InitEntity() *Location {
	entity := Location{
		VisitIdsMap: make(map[uint32]bool)}
	return &entity
}

func (repo *LocationsRepo) Create(data *JsonData) bool {
	entity := repo.InitEntity()
	ok := entity.Update(data, false)
	if !ok {
		return false
	}
	repo.Add(entity)
	return true
}

// func (repo *LocationsRepo) CreateFromJson(raw_data []byte) error {
// 	entity := repo.InitEntity()
// 	err := json.Unmarshal(raw_data, entity)
// 	if err == nil {
// 		repo.Add(entity)
// 	}
// 	return err
// }

func (repo *LocationsRepo) Add(entity *Location) {
	repo.Mutex.Lock()
	repo.Collection[entity.Id] = entity
	repo.Mutex.Unlock()
}

func (repo *LocationsRepo) Find(id uint32) (*Location, bool) {
	repo.Mutex.RLock()
	var entity, found = repo.Collection[id]
	repo.Mutex.RUnlock()
	return entity, found
}

func (repo *LocationsRepo) FindEntity(id uint32) (Entity, bool) {
	return repo.Find(id)
}
