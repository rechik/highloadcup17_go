package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
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

func (entity *Location) Update(data *JsonData, lock bool) {
	if lock {
		entity.Mutex.Lock()
	}
	for key, value := range *data {
		switch key {
		case "id":
			entity.Id = uint32(value.(float64))
		case "place":
			entity.Place = value.(string)
		case "country":
			entity.Country = value.(string)
		case "city":
			entity.City = value.(string)
		case "distance":
			entity.Distance = uint32(value.(float64))
		}
	}
	//TODO: denormolize in Visits
	if lock {
		entity.Mutex.Unlock()
	}
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

func (entity *Location) WriteAvgsJson(w io.Writer, fromDate *uint32, toDate *uint32, fromAge *uint32, toAge *uint32, gender *string) {
	entity.Mutex.RLock()

	visits, _ := Visits.FindAll(entity.VisitIds())

	marks_count := 0
	marks_sum := uint32(0)
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
		// TODO: fromAge, toAge
		if gender != nil && visit.UserGender != *gender {
			visit.Mutex.RUnlock()
			continue
		}
		marks_sum += visit.Mark
		marks_count += 1
		visit.Mutex.RUnlock()
	}
	entity.Mutex.RUnlock()

	avg_str := "0"
	if marks_count > 0 {
		avg := float64(marks_sum) / float64(marks_count)
		avg_str = fmt.Sprintf("%.5f", avg)
	}

	w.Write([]byte("{\"avg\": "))
	w.Write([]byte(avg_str))
	w.Write([]byte("}"))
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

func (repo *LocationsRepo) Create(data *JsonData) {
	entity := repo.InitEntity()
	entity.Update(data, false)
	repo.Add(entity)
}

func (repo *LocationsRepo) CreateFromJson(raw_data []byte) error {
	entity := repo.InitEntity()
	err := json.Unmarshal(raw_data, entity)
	if err == nil {
		repo.Add(entity)
	}
	return err
}

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
