package routes

import (
	"cosmetics/routes/utils"
	"cosmetics/util"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

const createQuery = `
	insert into cosmetics(id, version, data) values($1, $2, $3)
`

func CreateCosmetic(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	var data = make(map[string]interface{})
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	util.Log("Trying to create cosmetic: ", data)
	var key = data["id"]
	var version = data["version"]
	_, ok := version.(float64)
	if key == nil || version == nil || !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	jsonData, _ := json.Marshal(data)
	result, err := ctx.Pool.Exec(ctx.Context, createQuery, key, version, jsonData)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				res.WriteHeader(http.StatusBadRequest)
			default:
				util.PrintData(result)
				util.PrintData(err)
				res.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			util.PrintData(result)
			util.PrintData(err)
			res.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	util.Log("Created cosmetic", "")

	res.WriteHeader(http.StatusOK)
}

const updateQuery = `
	update cosmetics set version = $2, data = $3 where id = $1
`

func UpdateCosmetic(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	var data = make(map[string]interface{})
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	util.Log("Trying to update cosmetic: ", data)
	var key = data["id"]
	var version = data["version"]
	_, ok := version.(float64)
	if key == nil || version == nil || !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	jsonData, _ := json.Marshal(data)
	result, err := ctx.Pool.Exec(ctx.Context, updateQuery, key, version, jsonData)
	if err != nil {
		util.PrintData(result)
		util.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() != 1 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	util.Log("Updated cosmetic", result.RowsAffected())

	res.WriteHeader(http.StatusOK)
}
