package main

import (
	"database/sql"
	"fmt"

	"github.com/xingmang/xingmang-desktop/internal/agent"
	_ "modernc.org/sqlite"
)

func openDB() (*sql.DB, error) {
	return sql.Open("sqlite", `D:////Temp////xm-test////data////xingmang.db`+"?_journal_mode=WAL&_busy_timeout=10000&_txlock=immediate")
}

func main() {
	db, err := openDB()
	if err != nil {
		panic(err)
	}
	m := agent.NewManager(db)
	events, err := m.EventsAfter("e80c243327f47c048609b9495ae95a12", 0)
	fmt.Println("err:", err)
	for _, e := range events {
		fmt.Printf("seq=%d type=%s payload=%s\n", e.Seq, e.Type, string(e.Payload)[:min(60, len(e.Payload))])
	}
}
