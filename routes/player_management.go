package routes

import (
	"cosmetics/internal"
	"cosmetics/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const createPlayer = `
	insert into players(id) values($1) on conflict do nothing;
`

const addPlayerCosmetic = `
	insert into player_cosmetics (player_id, cosmetic_id)  values($1, $2) on conflict do nothing;
`

func AddPlayerCosmetic(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")
	cosmeticId := req.PathValue("cosmetic_id")
	utils.LogData{
		Message: "Trying to add cosmetic to player!",
		Data: struct {
			Player   string
			Cosmetic string
		}{playerId, cosmeticId},
	}.Log()
	if !utils.IsValidResourceLocationNamespace(cosmeticId) {
		utils.LogData{
			Message: "Failed to add cosmetic, invalid cosmetic id!",
			Data: struct {
				Player   string
				Cosmetic string
			}{playerId, cosmeticId},
		}.Log()
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	_, _ = ctx.Pool.Exec(ctx.Context, createPlayer, playerId)
	result, err := ctx.Pool.Exec(ctx.Context, addPlayerCosmetic, playerId, cosmeticId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				res.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(res, "No matching cosmetic found!")
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
		utils.LogData{
			Message: "Failed to add player cosmetic!",
			Data:    err,
		}.Log()
		return
	}
	if result.RowsAffected() != 1 {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(res, "Already present!")
		utils.LogData{
			Message: "Failed to add cosmetic to player, no matching found!",
			Data: struct {
				Player   string
				Cosmetic string
			}{playerId, cosmeticId},
		}.Log()
		return
	}
	utils.LogData{
		Message: "Added cosmetic to player!",
		Data: struct {
			Player   string
			Cosmetic string
		}{playerId, cosmeticId},
	}.Log()
}

const removePlayerCosmetic = `
	delete from player_cosmetics where player_id = $1 and cosmetic_id = $2
`

func RemovePlayerCosmetic(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")
	cosmeticId := req.PathValue("cosmetic_id")
	utils.LogData{
		Message: "Trying to remove cosmetic from player!",
		Data: struct {
			Player   string
			Cosmetic string
		}{playerId, cosmeticId},
	}.Log()
	if !utils.IsValidResourceLocationNamespace(cosmeticId) {
		utils.LogData{
			Message: "Failed to remove cosmetic from player, invalid id!",
			Data: struct {
				Player   string
				Cosmetic string
			}{playerId, cosmeticId},
		}.Log()
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := ctx.Pool.Exec(ctx.Context, removePlayerCosmetic, playerId, cosmeticId)
	if err != nil {
		utils.PrintData(result)
		utils.PrintData(err)
		utils.LogData{
			Message: "Failed to remove cosmetic from player!",
			Data:    err,
		}.Log()
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() != 1 {
		utils.LogData{
			Message: "Failed to remove cosmetic from player, no matching found!",
			Data: struct {
				Player   string
				Cosmetic string
			}{playerId, cosmeticId},
		}.Log()
		res.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(res, "No matching pair found!")
		return
	}
}

const setPlayerCustomData = `
	insert into players (id, data) values ($1, $2) on conflict (id) do update set data = excluded.data;
`

func UpdatePlayerCustomData(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	data := string(body)
	if !utils.IsJson(data) {
		utils.LogData{
			Message: "Failed Updating player, invalid json!",
			Data: struct {
				Player string
				Data   string
			}{playerId, string(body)},
		}.Log()
		res.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(res, "Invalid json!")
		return
	}

	utils.LogData{
		Message: "Update custom data for player",
		Data: struct {
			Player string
			Data   string
		}{playerId, data},
	}.Log()

	_, err = ctx.Pool.Exec(ctx.Context, setPlayerCustomData, playerId, data)
	if err != nil {
		utils.LogData{
			Message: "Failed to update custom player data",
			Data: struct {
				Error error
				Data  string
			}{err, data},
		}.Log()
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

const getPlayerCustomData = `
	select data from players where id = $1
`

func GetPlayerCustomData(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")
	result, err := ctx.Pool.Query(ctx.Context, getPlayerCustomData, playerId)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer result.Close()
	if !result.Next() {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, _ = res.Write(result.RawValues()[0])
}

const deletePlayerQuery = `
	delete from players where id = $1
`

func DeletePlayer(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")

	utils.LogData{
		Message: "Trying to delete player",
		Data:    playerId,
	}.Log()
	result, err := ctx.Pool.Exec(ctx.Context, deletePlayerQuery, playerId)
	if err != nil {
		utils.LogData{
			Message: "Failed to delete player!",
			Data:    err,
		}.Log()
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	if result.RowsAffected() != 1 {
		utils.LogData{
			Message: "Failed to delete player, no matching found!",
			Data:    playerId,
		}.Log()
		res.WriteHeader(http.StatusNotFound)
	}
}

const getPlayerQuery = `
	select player_id, player_data, cosmetics from players_with_cosmetics where player_id = $1 limit 1
`

func GetPlayerData(ctx internal.RouteContext, res http.ResponseWriter, req *http.Request) {
	playerId := req.PathValue("uuid")
	result, err := ctx.Pool.Query(ctx.Context, getPlayerQuery, playerId)

	if err != nil {
		utils.PrintData(result)
		utils.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer result.Close()
	player, err := pgx.CollectOneRow(result, pgx.RowToStructByPos[PlayerType])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		utils.PrintData(result)
		utils.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(player)
	if err != nil {
		utils.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = io.WriteString(res, string(data))
}

const getPlayerIds = `
	select id from players
`

func ListPlayerIds(ctx internal.RouteContext, res http.ResponseWriter, _ *http.Request) {
	var players, err = ctx.Pool.Query(ctx.Context, getPlayerIds)
	if err != nil {
		utils.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer players.Close()

	var list = make([]string, 0)
	for players.Next() {
		id, err := uuid.FromBytes(players.RawValues()[0])
		if err != nil {
			utils.LogData{
				Message: "Failed to create uuid",
				Data:    err,
			}.Log()
			continue
		}
		list = append(list, id.String())
	}

	data, err := json.Marshal(list)
	if err != nil {
		utils.PrintData(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = io.WriteString(res, string(data))
}
