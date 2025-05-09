// Code generated by ent, DO NOT EDIT.

package mlscore

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the mlscore type in the database.
	Label = "ml_score"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldUserID holds the string denoting the user_id field in the database.
	FieldUserID = "user_id"
	// FieldAdvertiserID holds the string denoting the advertiser_id field in the database.
	FieldAdvertiserID = "advertiser_id"
	// FieldScore holds the string denoting the score field in the database.
	FieldScore = "score"
	// EdgeUser holds the string denoting the user edge name in mutations.
	EdgeUser = "user"
	// EdgeAdvertiser holds the string denoting the advertiser edge name in mutations.
	EdgeAdvertiser = "advertiser"
	// Table holds the table name of the mlscore in the database.
	Table = "ml_scores"
	// UserTable is the table that holds the user relation/edge.
	UserTable = "ml_scores"
	// UserInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UserInverseTable = "users"
	// UserColumn is the table column denoting the user relation/edge.
	UserColumn = "user_id"
	// AdvertiserTable is the table that holds the advertiser relation/edge.
	AdvertiserTable = "ml_scores"
	// AdvertiserInverseTable is the table name for the Advertiser entity.
	// It exists in this package in order to avoid circular dependency with the "advertiser" package.
	AdvertiserInverseTable = "advertisers"
	// AdvertiserColumn is the table column denoting the advertiser relation/edge.
	AdvertiserColumn = "advertiser_id"
)

// Columns holds all SQL columns for mlscore fields.
var Columns = []string{
	FieldID,
	FieldUserID,
	FieldAdvertiserID,
	FieldScore,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

// OrderOption defines the ordering options for the MlScore queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByUserID orders the results by the user_id field.
func ByUserID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUserID, opts...).ToFunc()
}

// ByAdvertiserID orders the results by the advertiser_id field.
func ByAdvertiserID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAdvertiserID, opts...).ToFunc()
}

// ByScore orders the results by the score field.
func ByScore(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldScore, opts...).ToFunc()
}

// ByUserField orders the results by user field.
func ByUserField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newUserStep(), sql.OrderByField(field, opts...))
	}
}

// ByAdvertiserField orders the results by advertiser field.
func ByAdvertiserField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAdvertiserStep(), sql.OrderByField(field, opts...))
	}
}
func newUserStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(UserInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, UserTable, UserColumn),
	)
}
func newAdvertiserStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(AdvertiserInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, AdvertiserTable, AdvertiserColumn),
	)
}
