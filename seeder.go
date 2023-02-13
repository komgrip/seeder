package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

type Seed struct {
	DB      *sql.DB
	Files   []fs.DirEntry
	SeedDir string
}

func (s Seed) ExecuteSeedData() error {
	for _, file := range s.Files {
		//get table substring from file's name
		tableName := file.Name()[5 : len(file.Name())-10]

		//get only query from seed file
		insertQueries, err := os.ReadFile(s.SeedDir + file.Name())
		if err != nil {
			return err
		}

		onlyInsertQueriesSlice := []string{}
		for _, query := range strings.Split(string(insertQueries), "\n") {
			if query != "" && query[0:2] != "--" {
				onlyInsertQueriesSlice = append(onlyInsertQueriesSlice, query)
			}
		}

		//join insert query
		insert := strings.Join(onlyInsertQueriesSlice, "\n")

		//check if has table in database
		_, err = s.DB.Query(fmt.Sprintf("select * from %s;", tableName))
		if err != nil {
			fmt.Println(err)
			continue
		}

		//check wheter table has data
		result, err := s.DB.Exec(fmt.Sprintf("select * from %s limit 1;", tableName))
		if err != nil {
			return err
		}

		rowAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		//reset nextval in postgres
		_, err = s.DB.Exec(fmt.Sprintf("alter sequence %s_id_seq restart with 1", tableName))
		if err != nil {
			log.Println(err)
			// return err
		}

		//insert seed data to database
		if rowAffected == 0 {
			_, err := s.DB.Exec(insert)
			if err != nil {
				log.Println(err)
				// return err
			}
			continue
		}

		_, err = s.DB.Exec(fmt.Sprintf("delete from %s;", tableName))
		if err != nil {
			log.Println(err)
			// return err
		}

		_, err = s.DB.Exec(insert)
		if err != nil {
			log.Println(err)
			// return err
		}
	}

	return nil
}

func GetSeedFiles(filesWantToSeed []string, seedDir string) ([]fs.DirEntry, error) {
	if len(filesWantToSeed) == 0 {
		files, err := os.ReadDir(seedDir)
		if err != nil {
			return []fs.DirEntry{}, err
		}

		return files, nil
	}

	result := []fs.DirEntry{}

	files, err := os.ReadDir(seedDir)
	if err != nil {
		return []fs.DirEntry{}, err
	}

	for _, file := range files {
		if slices.Contains(filesWantToSeed, file.Name()) {
			result = append(result, file)
		}
	}

	return result, nil
}

func CreateSeedFile(path string) error {
	// create file from given input
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	comment := "-- example: INSERT INTO table (column1, column2) VALUES (value1, value2);"
	_, err = file.WriteString(comment)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:        "seeder",
		Usage:       "use seeder to seed data to your database",
		Description: "seeder will help you seed your data to a table that have been migration to a database easily.",
		Version:     "1.0.0",
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "use to create seed file which is .sql file from given directory and table name",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "dir",
						Aliases:  []string{"d"},
						Value:    "",
						Usage:    "receive the path of directory, example: db/seeds",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					// check correctness of dir flag
					if string(ctx.String("dir")[0]) == "/" {
						return errors.New("error: directory's path is invalid")
					}

					//check if table name must be input
					if len(ctx.Args().Slice()) == 0 {
						return errors.New("error: table name argument is required")
					}

					//find the root of project directory
					workingDir, err := os.Getwd()
					if err != nil {
						return err
					}

					for _, tableName := range ctx.Args().Slice() {
						dirPath := workingDir + "/" + ctx.String("dir") + "/"
						seedFileName := fmt.Sprintf("seed_%s_table.sql", tableName)
						seedFilePath := dirPath + seedFileName

						//check whether dir is exist
						_, err = os.Stat(ctx.String("dir"))
						if os.IsNotExist(err) {
							//create dir with given path from cli's input
							err := os.MkdirAll(dirPath, os.ModePerm)
							if err != nil {
								return err
							}

							//create file and insert to previous created dir
							err = CreateSeedFile(seedFilePath)
							if err != nil {
								return err
							}

							continue
						}

						files, err := os.ReadDir(dirPath)
						if err != nil {
							return err
						}

						//check whether file duplicate or not
						for _, file := range files {
							if file.Name() == seedFileName {
								return errors.New("error: file has existed in directory")
							}
						}

						err = CreateSeedFile(seedFilePath)
						if err != nil {
							return err
						}
					}

					return nil
				},
			},
			{
				Name:  "seed",
				Usage: "use to insert data from given files to your database.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "database",
						Aliases:  []string{"db"},
						Value:    "",
						Usage:    "receive database url",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Value:    "",
						Usage:    "receive the path of directory that contains seed files",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					startTime := time.Now()

					//connect to database
					db, err := sql.Open("postgres", ctx.String("database"))
					if err != nil {
						return err
					}
					defer db.Close()

					//read file from given path
					workingDir, err := os.Getwd()
					if err != nil {
						return err
					}

					//find the root of project directory
					seedDir := workingDir + "/" + ctx.String("path") + "/"

					files, err := GetSeedFiles(ctx.Args().Slice(), seedDir)
					if err != nil {
						return err
					}

					if len(files) == 0 {
						return errors.New("error: file has not existed")
					}

					seed := Seed{DB: db, Files: files, SeedDir: seedDir}
					err = seed.ExecuteSeedData()
					if err != nil {
						return err
					}

					log.Printf("seeding time: %s\n", time.Since(startTime))

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
