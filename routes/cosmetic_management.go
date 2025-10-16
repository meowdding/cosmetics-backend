package routes

import (
	"cosmetics/internal"
	"cosmetics/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

const createQuery = `
	insert into cosmetics(id, version, data) values($1, $2, $3) on conflict (id) do update set version = $2, data = $3
`

func CreateOrUpdateCosmetic(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	cosmeticId := req.PathValue("cosmetic_id")
	var data = make(map[string]interface{})
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	data["id"] = cosmeticId
	utils.Log("Trying to create cosmetic: ", data)
	if cosmeticId == "" || !utils.IsValidResourceLocationNamespace(cosmeticId) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var version = data["version"]
	_, ok := version.(float64)
	if version == nil || !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	jsonData, _ := json.Marshal(data)
	result, err := ctx.Pool.Exec(ctx.Context, createQuery, cosmeticId, version, jsonData)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				res.WriteHeader(http.StatusBadRequest)
			default:
				utils.PrintData(result)
				utils.PrintData(err)
				res.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			utils.PrintData(result)
			utils.PrintData(err)
			res.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	utils.Log("Created cosmetic", "")

	res.WriteHeader(http.StatusOK)
}

const getQuery = `
	select data from cosmetics where id = $1
`

func GetCosmetic(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	cosmeticId := req.PathValue("cosmetic_id")
	if !utils.IsValidResourceLocationNamespace(cosmeticId) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := ctx.Pool.Query(ctx.Context, getQuery, cosmeticId)
	if err != nil {
		utils.PrintData(result)
		utils.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !result.Next() {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, _ = res.Write(result.RawValues()[0])
}

const deleteQuery = `
	delete from cosmetics where id = $1
`

func DeleteCosmetic(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	cosmeticId := req.PathValue("cosmetic_id")
	if !utils.IsValidResourceLocationNamespace(cosmeticId) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := ctx.Pool.Exec(ctx.Context, deleteQuery, cosmeticId)
	if err != nil {
		utils.PrintData(result)
		utils.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() != 1 {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	utils.Log("Deleted cosmetic", result.RowsAffected())

	res.WriteHeader(http.StatusOK)
}

const getCosmeticIds = `
	select id from cosmetics
`

func ListCosmeticIds(ctx internal.RouteContext, res http.ResponseWriter, _ *http.Request) {
	var cosmetics, err = ctx.Pool.Query(ctx.Context, getCosmeticIds)
	if err != nil {
		utils.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var list = make([]string, 0)
	for cosmetics.Next() {
		values, err := cosmetics.Values()
		if err != nil {
			continue
		}

		list = append(list, fmt.Sprintf("%v", values[0]))
	}

	data, err := json.Marshal(list)
	if err != nil {
		utils.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = io.WriteString(res, string(data))
}
