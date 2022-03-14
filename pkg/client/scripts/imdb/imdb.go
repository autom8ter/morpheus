package imdb

import (
	"context"
	"github.com/autom8ter/morpheus/pkg/client"
	"github.com/autom8ter/morpheus/pkg/client/scripts"
	"github.com/autom8ter/morpheus/pkg/graph/model"
	"github.com/jmoiron/sqlx"
	"github.com/palantir/stacktrace"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

const addRelation = `
query ($key: Key!, $direction: Direction!, $relationship: String!, $nodeKey: Key!) {
  get(key: $key){
    id
    addRelationship(direction: $direction, relationship: $relationship, nodeKey: $nodeKey) {
      type
    }
  }
}`

func ImportIMDB(db *sqlx.DB) scripts.Script {

	return func(ctx context.Context, client *client.Client) error {
		egp := errgroup.Group{}
		egp.Go(func() error {
			return syncActors(ctx, db, client)
		})
		egp.Go(func() error {
			return syncMovies(ctx, db, client)
		})
		egp.Go(func() error {
			return syncDirectors(ctx, db, client)
		})
		if err := egp.Wait(); err != nil {
			return stacktrace.Propagate(err, "")
		}
		egp2 := errgroup.Group{}
		// second group
		egp2.Go(func() error {
			return syncRoles(ctx, db, client)
		})
		//egp2.Go(func() error {
		//	return syncMovieGenres(ctx, db, client)
		//})
		if err := egp2.Wait(); err != nil {
			return stacktrace.Propagate(err, "")
		}
		egp3 := errgroup.Group{}
		egp3.Go(func() error {
			return syncMovieDirectors(ctx, db, client)
		})
		//egp3.Go(func() error {
		//	return syncDirectorGenres(ctx, db, client)
		//})
		return egp3.Wait()
	}
}

func syncActors(ctx context.Context, db *sqlx.DB, client *client.Client) error {
	actors, err := scripts.GetFullTable(db, "actors")
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	const addActor = `
query ($id: String, $first_name: String!, $last_name: String!, $gender: String!) {
    add(add: {
        id: $id,
        type: "actor",
        properties: {
            first_name: $first_name
			last_name: $last_name
			gender: $gender
        }
    }){
        id
    }
}
`
	for _, actor := range actors {
		_, err := client.Queryx(ctx, addActor, map[string]interface{}{
			"id":         cast.ToString(actor["id"]),
			"first_name": cast.ToString(actor["first_name"]),
			"last_name":  cast.ToString(actor["last_name"]),
			"gender":     cast.ToString(actor["gender"]),
		})
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
	}
	return nil
}

func syncMovies(ctx context.Context, db *sqlx.DB, client *client.Client) error {
	movies, err := scripts.GetFullTable(db, "movies")
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	const addMovie = `
query ($id: String, $name: String!, $year: String!, $rank: String!) {
    add(add: {
        id: $id,
        type: "movie",
        properties: {
            name: $name
			year: $year
			rank: $rank
        }
    }){
        id
    }
}
`
	for _, movie := range movies {
		_, err := client.Queryx(ctx, addMovie, map[string]interface{}{
			"id":   cast.ToString(movie["id"]),
			"name": cast.ToString(movie["name"]),
			"year": cast.ToString(movie["year"]),
			"rank": cast.ToString(movie["rank"]),
		})
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
	}
	return nil
}

func syncDirectors(ctx context.Context, db *sqlx.DB, client *client.Client) error {
	directors, err := scripts.GetFullTable(db, "directors")
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	const addDirector = `
query ($id: String, $first_name: String!, $last_name: String!) {
    add(add: {
        id: $id,
        type: "director",
        properties: {
            first_name: $first_name
			last_name: $last_name
        }
    }){
        id
    }
}
`
	for _, director := range directors {
		_, err := client.Queryx(ctx, addDirector, map[string]interface{}{
			"id":         cast.ToString(director["id"]),
			"first_name": cast.ToString(director["first_name"]),
			"last_name":  cast.ToString(director["last_name"]),
		})
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
	}
	return nil
}

func syncRoles(ctx context.Context, db *sqlx.DB, client *client.Client) error {
	roles, err := scripts.GetFullTable(db, "roles")
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	const addRole = `
query ($role: String) {
   add(add: {
       id: $role,
       type: "role",
   }){
       id
   }
}
`
	var haveAdded = map[string]struct{}{}
	for _, role := range roles {
		rolename := cast.ToString(role["role"])
		if _, ok := haveAdded[rolename]; !ok {
			// add nodes
			_, err := client.Queryx(ctx, addRole, map[string]interface{}{
				"role": rolename,
			})
			if err != nil {
				return stacktrace.Propagate(err, "")
			}
			haveAdded[rolename] = struct{}{}
		}

		_, err := client.Queryx(ctx, addRelation, map[string]interface{}{
			"key": &model.Key{
				Type: "actor",
				ID:   cast.ToString(role["actor_id"]),
			},
			"direction":    model.DirectionOutgoing,
			"relationship": "acted_in",
			"nodeKey": &model.Key{
				Type: "movie",
				ID:   cast.ToString(role["movie_id"]),
			},
		})
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
		_, err = client.Queryx(ctx, addRelation, map[string]interface{}{
			"key": &model.Key{
				Type: "actor",
				ID:   cast.ToString(role["actor_id"]),
			},
			"direction":    model.DirectionOutgoing,
			"relationship": "has_role",
			"nodeKey": &model.Key{
				Type: "role",
				ID:   rolename,
			},
		})
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
		_, err = client.Queryx(ctx, addRelation, map[string]interface{}{
			"key": &model.Key{
				Type: "movie",
				ID:   cast.ToString(role["movie_id"]),
			},
			"direction":    model.DirectionOutgoing,
			"relationship": "has_role",
			"nodeKey": &model.Key{
				Type: "role",
				ID:   rolename,
			},
		})
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
	}
	return nil
}

func syncMovieGenres(ctx context.Context, db *sqlx.DB, client *client.Client) error {
	movies_genres, err := scripts.GetFullTable(db, "movies_genres")
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	const addGenre = `
query ($genre: String) {
   add(add: {
       id: $genre,
       type: "genre"
   }){
       id
   }
}
`
	var haveAdded = map[string]struct{}{}
	for _, genre := range movies_genres {
		genreName := cast.ToString(genre["genre"])
		if _, ok := haveAdded[genreName]; !ok {
			_, err := client.Queryx(ctx, addGenre, map[string]interface{}{
				"genre": genreName,
			})
			if err != nil {
				return stacktrace.Propagate(err, "")
			}
			haveAdded[genreName] = struct{}{}
		}
		_, err := client.Queryx(ctx, addRelation, map[string]interface{}{
			"key": &model.Key{
				Type: "movie",
				ID:   cast.ToString(genre["movie_id"]),
			},
			"direction":    model.DirectionOutgoing,
			"relationship": "has_genre",
			"nodeKey": &model.Key{
				Type: "genre",
				ID:   genreName,
			},
		})
		if err != nil {
			return stacktrace.Propagate(err, genreName)
		}
	}
	return nil
}

func syncMovieDirectors(ctx context.Context, db *sqlx.DB, client *client.Client) error {
	movies_directors, err := scripts.GetFullTable(db, "movies_directors")
	if err != nil {
		return stacktrace.Propagate(err, "")
	}

	for _, mdirector := range movies_directors {
		_, err := client.Queryx(ctx, addRelation, map[string]interface{}{
			"key": &model.Key{
				Type: "director",
				ID:   cast.ToString(mdirector["director_id"]),
			},
			"direction":    model.DirectionOutgoing,
			"relationship": "directed",
			"nodeKey": &model.Key{
				Type: "movie",
				ID:   cast.ToString(mdirector["movie_id"]),
			},
		})
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
	}
	return nil
}

func syncDirectorGenres(ctx context.Context, db *sqlx.DB, client *client.Client) error {
	directors_genres, err := scripts.GetFullTable(db, "directors_genres")
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	for _, genre := range directors_genres {
		genreName := cast.ToString(genre["genre"])

		_, err := client.Queryx(ctx, addRelation, map[string]interface{}{
			"key": &model.Key{
				Type: "director",
				ID:   cast.ToString(genre["director_id"]),
			},
			"direction":    model.DirectionIncoming,
			"relationship": "has_genre",
			"nodeKey": &model.Key{
				Type: "genre",
				ID:   genreName,
			},
		})
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
	}
	return nil
}
