// Copyright 2022 The Matrix.org Foundation C.I.C.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/matrix-org/dendrite/internal/sqlutil"
	"github.com/matrix-org/dendrite/syncapi/storage/tables"
	"github.com/matrix-org/gomatrixserverlib"
)

const countUsersLastSeenAfterSQL = ""+
	"SELECT COUNT(*) FROM ("+
	" SELECT user_id FROM device_devices WHERE last_seen > $1 "+
	" GROUP BY user_id"+
	" )"

const countActiveRoomsSQL = ""+
	"SELECT COUNT(DISTINCT room_id) FROM syncapi_output_room_events"+
	" WHERE type = $1 AND id > $2"

type statsStatements struct {
	serverName string
	countUsersLastSeenAfterStmt *sql.Stmt
	countActiveRoomsStmt *sql.Stmt
}

func PrepareStats(db *sql.DB, serverName string) (tables.Stats, error) {
	s := &statsStatements{
		serverName: serverName,
	}
	return s, sqlutil.StatementList{
		{&s.countUsersLastSeenAfterStmt, countUsersLastSeenAfterSQL},
		{&s.countActiveRoomsStmt, countActiveRoomsSQL},
	}.Prepare(db)
}

func (s *statsStatements) DailyUsers(ctx context.Context, txn *sql.Tx) (result int64, err error) {
	stmt := sqlutil.TxStmt(txn, s.countUsersLastSeenAfterStmt)
	lastSeenAfter := time.Now().AddDate(0, 0, 1)
	err = stmt.QueryRowContext(ctx,
		gomatrixserverlib.AsTimestamp(lastSeenAfter),
	).Scan(&result)
	return
}

func (s *statsStatements) MonthlyUsers(ctx context.Context, txn *sql.Tx) (result int64, err error) {
	stmt := sqlutil.TxStmt(txn, s.countUsersLastSeenAfterStmt)
	lastSeenAfter := time.Now().AddDate(0, 0, 30)
	err = stmt.QueryRowContext(ctx,
		gomatrixserverlib.AsTimestamp(lastSeenAfter),
	).Scan(&result)
	return
}