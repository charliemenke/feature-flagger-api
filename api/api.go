package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type FeatureFlaggerAPI struct {
	router *mux.Router
	redis  *redis.Client
}

type Feature struct {
	Name    *string `json:"name,omitempty"`
	Enabled *bool   `json:"enabled"`
}

type UpdateRequest struct {
	Enabled *bool `json:"enabled,omitempty"`
}

func (api *FeatureFlaggerAPI) Initialize(redisHost string, redisPort string, redisID int, redisPass string) {
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()

	// redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPass,
		DB:       redisID,
	})
	api.redis = redisClient

	// routes
	apiRouter.HandleFunc("/health-check", healthCheck).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/features", listFeaturesHandler(api.redis)).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/features", createFeaturesHandler(api.redis)).Methods("POST", "OPTIONS")

	apiRouter.HandleFunc("/features/{key}", getFeatureHandler(api.redis)).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/features/{key}", updateFeatureHandler(api.redis)).Methods("PUT", "OPTIONS")
	apiRouter.HandleFunc("/features/{key}", deleteFeatureHandler(api.redis)).Methods("DELETE", "OPTIONS")

	api.router = router
}

func (api *FeatureFlaggerAPI) Start(port string) {
	log.Infof("Feature Flagger API listing on port %s\n", port)
	err := http.ListenAndServe(":"+port, api.router)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Errorln("API Error")
	}
}

func listFeaturesHandler(redisDB *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		keys, err := redisDB.Keys("*").Result()
		if err != nil {
			log.Errorf("Error getting features from redis: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error getting features: " + err.Error()))
			return
		}
		var features []Feature
		for _, key := range keys {
			enabled, err := redisDB.Get(key).Result()
			if err != nil {
				log.Errorf("Failed to get value for feature key\n")
			}
			enabledBool := true
			if enabled != "1" {
				enabledBool = false
			}
			keyVal := key
			features = append(features, Feature{Name: &keyVal, Enabled: &enabledBool})
		}
		featuresJSON, err := json.Marshal(features)
		if err != nil {
			log.Errorf("Error marsheling features to JSON: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error preparing features to be returned: " + err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(featuresJSON)
	}
}
func createFeaturesHandler(redisDB *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		// ready request body
		var body Feature
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			log.Errorf("Failed to decode request body\n")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error reading request: " + err.Error()))
			return
		}
		// if no feature name specifiged, return error
		if body.Name == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("You must specify feature name."))
			return
		}
		if body.Enabled == nil {
			// default to false
			*body.Enabled = false
		}
		// set feature in redis "featurename: enabled"
		_, err = redisDB.Get(*body.Name).Result()
		if err != nil && err == redis.Nil {
			err = redisDB.Set(*body.Name, *body.Enabled, 0).Err()
			if err != nil {
				log.Errorf("Error adding feature to redis: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error adding feature: " + err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Succesfully created feature " + *body.Name))
			return
		} else if err != nil && err != redis.Nil {
			log.Errorf("Error checking if feature already exists: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error checking if feature already exists"))
			return
		} else {
			log.Errorf("Not creating, feature already exists: %s\n", err)
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("This feature already exists, please update or delete it instead"))
			return
		}
	}
}

func getFeatureHandler(redisDB *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		// get feature name to retrieve from url
		key := mux.Vars(r)["key"]
		enabled, err := redisDB.Get(key).Result()
		if err != nil {
			if err == redis.Nil {
				log.Errorf("No such feature found to update: %s\n", err)
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Feature: '" + key + "' does not exist in the database. Please check the feature name and try again."))
				return
			} else {
				log.Errorf("Error getting feature from redis: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error getting feature: " + err.Error()))
				return
			}
		}
		// covert bool to string
		enabledBool := true
		if enabled != "1" {
			enabledBool = false
		}
		// create and return feature struct
		feature := Feature{
			Name:    &key,
			Enabled: &enabledBool,
		}
		featureJSON, err := json.Marshal(feature)
		if err != nil {
			log.Errorf("Error marsheling body to JSON: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error preparing feature to be returned: " + err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(featureJSON)
	}
}

func updateFeatureHandler(redisDB *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		// decode body
		var body UpdateRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			log.Errorf("Failed to decode request body\n")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error reading request: " + err.Error()))
			return
		}
		// check that body has required key
		if body.Enabled == nil {
			log.Errorf("Body does not supply enabled field\n")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Request must supply 'enabled' field"))
			return
		}
		// get feature based on key
		key := mux.Vars(r)["key"]
		enabled, err := redisDB.Get(key).Result()
		if err != nil {
			if err == redis.Nil {
				log.Errorf("No such feature found to update: %s\n", err)
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Feature: '" + key + "' does not exist in the database. Please check the feature name and try again."))
				return
			} else {
				log.Errorf("Error getting feature from redis: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error getting feature: " + err.Error()))
				return
			}
		}
		// compare if we need to actualy update or not
		enabledBool := true
		if enabled != "1" {
			enabledBool = false
		}
		if *body.Enabled == enabledBool {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(key + " is already set to " + enabled))
			return
		}
		//update
		err = redisDB.Set(key, *body.Enabled, 0).Err()
		if err != nil {
			log.Errorf("Error updating feature in redis: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error updating feature: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Succesfully updated feature: " + key))
	}
}

func deleteFeatureHandler(redisDB *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		// get feature based on key
		key := mux.Vars(r)["key"]
		numDeleted, err := redisDB.Del(key).Result()
		if err != nil {
			log.Errorf("Error deleting feature from redis: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error deleting feature: " + err.Error()))
			return
		}
		if numDeleted <= 0 {
			w.WriteHeader(404)
			w.Write([]byte("Could not find feature to delete: " + key))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Succesfully deleted key: " + key))
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
}
