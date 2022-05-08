package postgres

import (
	"context"

	"darlinggo.co/pan"

	"lockbox.dev/clients"
)

func createSQL(_ context.Context, client Client) *pan.Query {
	return pan.Insert(client)
}

func getSQL(_ context.Context, id string) *pan.Query {
	var client Client
	q := pan.New("SELECT " + pan.Columns(client).String() + " FROM " + pan.Table(client))
	q.Where()
	q.Comparison(client, "ID", "=", id)
	return q.Flush(" ")
}

func listRedirectURIsSQL(_ context.Context, clientID string) *pan.Query {
	var redirectURI RedirectURI
	q := pan.New("SELECT " + pan.Columns(redirectURI).String() + " FROM " + pan.Table(redirectURI))
	q.Where()
	q.Comparison(redirectURI, "ClientID", "=", clientID)
	q.OrderByDesc("uri")
	return q.Flush(" ")
}

func updateSQL(_ context.Context, id string, change clients.Change) *pan.Query {
	var client Client
	query := pan.New("UPDATE " + pan.Table(client) + " SET ")
	if change.Name != nil {
		query.Assign(client, "Name", *change.Name)
	}
	if change.SecretHash != nil {
		query.Assign(client, "SecretHash", *change.SecretHash)
	}
	if change.SecretScheme != nil {
		query.Assign(client, "SecretScheme", *change.SecretScheme)
	}
	query.Flush(", ")
	query.Where()
	query.Comparison(client, "ID", "=", id)
	return query.Flush(" ")
}

func deleteSQL(_ context.Context, id string) *pan.Query {
	var client Client
	q := pan.New("DELETE FROM " + pan.Table(client))
	q.Where()
	q.Comparison(client, "ID", "=", id)
	return q.Flush(" ")
}

func addRedirectURIsSQL(_ context.Context, uris []RedirectURI) *pan.Query {
	tableNamers := make([]pan.SQLTableNamer, 0, len(uris))
	for _, uri := range uris {
		tableNamers = append(tableNamers, uri)
	}
	return pan.Insert(tableNamers...)
}

func removeRedirectURIsSQL(_ context.Context, uris []string) *pan.Query {
	var uri RedirectURI
	query := pan.New("DELETE FROM " + pan.Table(uri))
	query.Where()
	interfaces := make([]interface{}, 0, len(uris))
	for _, uri := range uris {
		interfaces = append(interfaces, uri)
	}
	query.In(uri, "ID", interfaces...)
	return query.Flush(" ")
}
