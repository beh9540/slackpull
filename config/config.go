package config

import (
        "sync/atomic"
        "database/sql"
        _ "github.com/go-sql-driver/mysql"
        "log"
        "sync"
)

var mt sync.Mutex

type Config struct {
        value atomic.Value
        SqlString string
}

type RepoConfig struct {
        Name string `json:"name"`
        Url string `json:"url"`
}

func (config *Config) Load() error {
        var (
                repoConfig map[string]string
                repo string
                webhook string
        )
        repoConfig = make(map[string]string)
        db, err := sql.Open("mysql", config.SqlString)
        if err != nil {
                return err
        }
        defer db.Close()

        rows, err := db.Query("select repo, url from webhook")
        if err != nil {
                log.Fatal(err)
        }
        defer rows.Close()
        for rows.Next() {
                err := rows.Scan(&repo, &webhook)
                if err != nil {
                        log.Fatal(err)
                }
                repoConfig[repo] = webhook
        }
        err = rows.Err()
        if err != nil {
                return err
        }
        config.value.Store(repoConfig)
        return nil
}

func (config *Config) Upsert(newConfig RepoConfig) error {
        mt.Lock()
        defer mt.Unlock()

        db, err := sql.Open("mysql", config.SqlString)
        if err != nil {
                return err
        }
        defer db.Close()

        _, err = db.Exec("INSERT INTO webhook (repo, url) VALUES(?, ?)", newConfig.Name, newConfig.Url)
        if err != nil {
                return err
        }

        configImpl := config.value.Load().(map[string]string)
        configImpl[newConfig.Name] = newConfig.Url
        config.value.Store(configImpl)
        return nil
}

func (config *Config) Get() map[string]string{
        return config.value.Load().(map[string]string)
}
