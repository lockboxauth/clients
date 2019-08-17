package postgres

import (
	"context"

	"darlinggo.co/pan"
	"lockbox.dev/clients"
)

func createSQL(ctx context.Context, client Client) *pan.Query {
	return pan.Insert(client)
}

func getSQL(ctx context.Context, id string) *pan.Query {
	var client Client
	q := pan.New("SELECT " + pan.Columns(client).String() + " FROM " + pan.Table(client))
	q.Where()
	q.Comparison(client, "ID", "=", id)
	return q.Flush(" ")
}

func listRedirectURIsSQL(ctx context.Context, clientID string) *pan.Query {
	var redirectURI RedirectURI
	q := pan.New("SELECT " + pan.Columns(redirectURI).String() + " FROM " + pan.Table(redirectURI))
	q.Where()
	q.Comparison(redirectURI, "ClientID", "=", clientID)
	q.OrderByDesc("uri")
	return q.Flush(" ")
}

func updateSQL(ctx context.Context, id string, change clients.Change) *pan.Query {
	var client Client
	q := pan.New("UPDATE " + pan.Table(client) + " SET ")
	if change.Name != nil {
		q.Assign(client, "Name", *change.Name)
	}
	if change.SecretHash != nil {
		q.Assign(client, "SecretHash", *change.SecretHash)
	}
	if change.SecretScheme != nil {
		q.Assign(client, "SecretScheme", *change.SecretScheme)
	}
	q.Flush(", ")
	q.Where()
	q.Comparison(client, "ID", "=", id)
	return q.Flush(" ")
}

func deleteSQL(ctx context.Context, id string) *pan.Query {
	var client Client
	q := pan.New("DELETE FROM " + pan.Table(client))
	q.Where()
	q.Comparison(client, "ID", "=", id)
	return q.Flush(" ")
}

func addRedirectURIsSQL(ctx context.Context, uris []RedirectURI) *pan.Query {
	tableNamers := make([]pan.SQLTableNamer, 0, len(uris))
	for _, uri := range uris {
		tableNamers = append(tableNamers, uri)
	}
	return pan.Insert(tableNamers...)
}

func removeRedirectURIsSQL(ctx context.Context, uris []string) *pan.Query {
	var uri RedirectURI
	q := pan.New("DELETE FROM " + pan.Table(uri))
	q.Where()
	interfaces := make([]interface{}, 0, len(uris))
	for _, uri := range uris {
		interfaces = append(interfaces, uri)
	}
	q.In(uri, "ID", interfaces...)
	return q.Flush(" ")
}
