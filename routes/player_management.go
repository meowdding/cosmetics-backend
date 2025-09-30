package routes

import (
	"cosmetics/routes/utils"
	"cosmetics/util"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const createPlayer = `
	insert into players(id) values($1) on conflict do nothing;
`

const addPlayerCosmetic = `
	insert into player_cosmetics (player_id, cosmetic_id)  values($1, $2) on conflict do nothing;
`

func AddPlayerCosmetic(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")
	cosmeticId := req.PathValue("cosmetic_id")
	if !util.IsValidResourceLocationNamespace(cosmeticId) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	_, _ = ctx.Pool.Exec(ctx.Context, createPlayer, playerId)
	result, err := ctx.Pool.Exec(ctx.Context, addPlayerCosmetic, playerId, cosmeticId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505", "23503":
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
}

const removePlayerCosmetic = `
	delete from player_cosmetics where player_id = $1 and cosmetic_id = $2
`

func RemovePlayerCosmetic(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")
	cosmeticId := req.PathValue("cosmetic_id")
	if !util.IsValidResourceLocationNamespace(cosmeticId) {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := ctx.Pool.Exec(ctx.Context, removePlayerCosmetic, playerId, cosmeticId)
	if err != nil {
		util.PrintData(result)
		util.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

const setPlayerCustomData = `
	insert into players (id, data) values ($1, $2) on conflict (id) do update set data = excluded.data;
`

func UpdatePlayerCustomData(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	data := string(body)

	result, err := ctx.Pool.Exec(ctx.Context, setPlayerCustomData, playerId, data)
	if err != nil {
		util.PrintData(result)
		util.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

const deletePlayerQuery = `
	delete from players where id = $1
`

func DeletePlayer(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")
	result, err := ctx.Pool.Exec(ctx.Context, deletePlayerQuery, playerId)
	if err != nil {
		util.PrintData(result)
		util.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

const getPlayerQuery = `
	select player_id, player_data, cosmetics from players_with_cosmetics where player_id = $1 limit 1
`

func GetPlayerData(ctx utils.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")
	result, err := ctx.Pool.Query(ctx.Context, getPlayerQuery, playerId)

	if err != nil {
		util.PrintData(result)
		util.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	player, err := pgx.CollectOneRow(result, pgx.RowToStructByPos[PlayerType])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		util.PrintData(result)
		util.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(player)
	if err != nil {
		util.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = io.WriteString(res, string(data))
}
