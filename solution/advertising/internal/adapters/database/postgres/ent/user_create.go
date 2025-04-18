// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"nlypage-final/internal/adapters/database/postgres/ent/user"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserCreate is the builder for creating a User entity.
type UserCreate struct {
	config
	mutation *UserMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetLogin sets the "login" field.
func (uc *UserCreate) SetLogin(s string) *UserCreate {
	uc.mutation.SetLogin(s)
	return uc
}

// SetAge sets the "age" field.
func (uc *UserCreate) SetAge(i int) *UserCreate {
	uc.mutation.SetAge(i)
	return uc
}

// SetNillableAge sets the "age" field if the given value is not nil.
func (uc *UserCreate) SetNillableAge(i *int) *UserCreate {
	if i != nil {
		uc.SetAge(*i)
	}
	return uc
}

// SetLocation sets the "location" field.
func (uc *UserCreate) SetLocation(s string) *UserCreate {
	uc.mutation.SetLocation(s)
	return uc
}

// SetNillableLocation sets the "location" field if the given value is not nil.
func (uc *UserCreate) SetNillableLocation(s *string) *UserCreate {
	if s != nil {
		uc.SetLocation(*s)
	}
	return uc
}

// SetGender sets the "gender" field.
func (uc *UserCreate) SetGender(u user.Gender) *UserCreate {
	uc.mutation.SetGender(u)
	return uc
}

// SetNillableGender sets the "gender" field if the given value is not nil.
func (uc *UserCreate) SetNillableGender(u *user.Gender) *UserCreate {
	if u != nil {
		uc.SetGender(*u)
	}
	return uc
}

// SetID sets the "id" field.
func (uc *UserCreate) SetID(u uuid.UUID) *UserCreate {
	uc.mutation.SetID(u)
	return uc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (uc *UserCreate) SetNillableID(u *uuid.UUID) *UserCreate {
	if u != nil {
		uc.SetID(*u)
	}
	return uc
}

// Mutation returns the UserMutation object of the builder.
func (uc *UserCreate) Mutation() *UserMutation {
	return uc.mutation
}

// Save creates the User in the database.
func (uc *UserCreate) Save(ctx context.Context) (*User, error) {
	uc.defaults()
	return withHooks(ctx, uc.sqlSave, uc.mutation, uc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (uc *UserCreate) SaveX(ctx context.Context) *User {
	v, err := uc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (uc *UserCreate) Exec(ctx context.Context) error {
	_, err := uc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (uc *UserCreate) ExecX(ctx context.Context) {
	if err := uc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (uc *UserCreate) defaults() {
	if _, ok := uc.mutation.ID(); !ok {
		v := user.DefaultID()
		uc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (uc *UserCreate) check() error {
	if _, ok := uc.mutation.Login(); !ok {
		return &ValidationError{Name: "login", err: errors.New(`ent: missing required field "User.login"`)}
	}
	if v, ok := uc.mutation.Age(); ok {
		if err := user.AgeValidator(v); err != nil {
			return &ValidationError{Name: "age", err: fmt.Errorf(`ent: validator failed for field "User.age": %w`, err)}
		}
	}
	if v, ok := uc.mutation.Gender(); ok {
		if err := user.GenderValidator(v); err != nil {
			return &ValidationError{Name: "gender", err: fmt.Errorf(`ent: validator failed for field "User.gender": %w`, err)}
		}
	}
	return nil
}

func (uc *UserCreate) sqlSave(ctx context.Context) (*User, error) {
	if err := uc.check(); err != nil {
		return nil, err
	}
	_node, _spec := uc.createSpec()
	if err := sqlgraph.CreateNode(ctx, uc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	uc.mutation.id = &_node.ID
	uc.mutation.done = true
	return _node, nil
}

func (uc *UserCreate) createSpec() (*User, *sqlgraph.CreateSpec) {
	var (
		_node = &User{config: uc.config}
		_spec = sqlgraph.NewCreateSpec(user.Table, sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = uc.conflict
	if id, ok := uc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := uc.mutation.Login(); ok {
		_spec.SetField(user.FieldLogin, field.TypeString, value)
		_node.Login = value
	}
	if value, ok := uc.mutation.Age(); ok {
		_spec.SetField(user.FieldAge, field.TypeInt, value)
		_node.Age = value
	}
	if value, ok := uc.mutation.Location(); ok {
		_spec.SetField(user.FieldLocation, field.TypeString, value)
		_node.Location = value
	}
	if value, ok := uc.mutation.Gender(); ok {
		_spec.SetField(user.FieldGender, field.TypeEnum, value)
		_node.Gender = value
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.User.Create().
//		SetLogin(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.UserUpsert) {
//			SetLogin(v+v).
//		}).
//		Exec(ctx)
func (uc *UserCreate) OnConflict(opts ...sql.ConflictOption) *UserUpsertOne {
	uc.conflict = opts
	return &UserUpsertOne{
		create: uc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.User.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (uc *UserCreate) OnConflictColumns(columns ...string) *UserUpsertOne {
	uc.conflict = append(uc.conflict, sql.ConflictColumns(columns...))
	return &UserUpsertOne{
		create: uc,
	}
}

type (
	// UserUpsertOne is the builder for "upsert"-ing
	//  one User node.
	UserUpsertOne struct {
		create *UserCreate
	}

	// UserUpsert is the "OnConflict" setter.
	UserUpsert struct {
		*sql.UpdateSet
	}
)

// SetLogin sets the "login" field.
func (u *UserUpsert) SetLogin(v string) *UserUpsert {
	u.Set(user.FieldLogin, v)
	return u
}

// UpdateLogin sets the "login" field to the value that was provided on create.
func (u *UserUpsert) UpdateLogin() *UserUpsert {
	u.SetExcluded(user.FieldLogin)
	return u
}

// SetAge sets the "age" field.
func (u *UserUpsert) SetAge(v int) *UserUpsert {
	u.Set(user.FieldAge, v)
	return u
}

// UpdateAge sets the "age" field to the value that was provided on create.
func (u *UserUpsert) UpdateAge() *UserUpsert {
	u.SetExcluded(user.FieldAge)
	return u
}

// AddAge adds v to the "age" field.
func (u *UserUpsert) AddAge(v int) *UserUpsert {
	u.Add(user.FieldAge, v)
	return u
}

// ClearAge clears the value of the "age" field.
func (u *UserUpsert) ClearAge() *UserUpsert {
	u.SetNull(user.FieldAge)
	return u
}

// SetLocation sets the "location" field.
func (u *UserUpsert) SetLocation(v string) *UserUpsert {
	u.Set(user.FieldLocation, v)
	return u
}

// UpdateLocation sets the "location" field to the value that was provided on create.
func (u *UserUpsert) UpdateLocation() *UserUpsert {
	u.SetExcluded(user.FieldLocation)
	return u
}

// ClearLocation clears the value of the "location" field.
func (u *UserUpsert) ClearLocation() *UserUpsert {
	u.SetNull(user.FieldLocation)
	return u
}

// SetGender sets the "gender" field.
func (u *UserUpsert) SetGender(v user.Gender) *UserUpsert {
	u.Set(user.FieldGender, v)
	return u
}

// UpdateGender sets the "gender" field to the value that was provided on create.
func (u *UserUpsert) UpdateGender() *UserUpsert {
	u.SetExcluded(user.FieldGender)
	return u
}

// ClearGender clears the value of the "gender" field.
func (u *UserUpsert) ClearGender() *UserUpsert {
	u.SetNull(user.FieldGender)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.User.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(user.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *UserUpsertOne) UpdateNewValues() *UserUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(user.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.User.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *UserUpsertOne) Ignore() *UserUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *UserUpsertOne) DoNothing() *UserUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the UserCreate.OnConflict
// documentation for more info.
func (u *UserUpsertOne) Update(set func(*UserUpsert)) *UserUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&UserUpsert{UpdateSet: update})
	}))
	return u
}

// SetLogin sets the "login" field.
func (u *UserUpsertOne) SetLogin(v string) *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.SetLogin(v)
	})
}

// UpdateLogin sets the "login" field to the value that was provided on create.
func (u *UserUpsertOne) UpdateLogin() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.UpdateLogin()
	})
}

// SetAge sets the "age" field.
func (u *UserUpsertOne) SetAge(v int) *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.SetAge(v)
	})
}

// AddAge adds v to the "age" field.
func (u *UserUpsertOne) AddAge(v int) *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.AddAge(v)
	})
}

// UpdateAge sets the "age" field to the value that was provided on create.
func (u *UserUpsertOne) UpdateAge() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.UpdateAge()
	})
}

// ClearAge clears the value of the "age" field.
func (u *UserUpsertOne) ClearAge() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.ClearAge()
	})
}

// SetLocation sets the "location" field.
func (u *UserUpsertOne) SetLocation(v string) *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.SetLocation(v)
	})
}

// UpdateLocation sets the "location" field to the value that was provided on create.
func (u *UserUpsertOne) UpdateLocation() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.UpdateLocation()
	})
}

// ClearLocation clears the value of the "location" field.
func (u *UserUpsertOne) ClearLocation() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.ClearLocation()
	})
}

// SetGender sets the "gender" field.
func (u *UserUpsertOne) SetGender(v user.Gender) *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.SetGender(v)
	})
}

// UpdateGender sets the "gender" field to the value that was provided on create.
func (u *UserUpsertOne) UpdateGender() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.UpdateGender()
	})
}

// ClearGender clears the value of the "gender" field.
func (u *UserUpsertOne) ClearGender() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.ClearGender()
	})
}

// Exec executes the query.
func (u *UserUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for UserCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *UserUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *UserUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: UserUpsertOne.ID is not supported by MySQL driver. Use UserUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *UserUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// UserCreateBulk is the builder for creating many User entities in bulk.
type UserCreateBulk struct {
	config
	err      error
	builders []*UserCreate
	conflict []sql.ConflictOption
}

// Save creates the User entities in the database.
func (ucb *UserCreateBulk) Save(ctx context.Context) ([]*User, error) {
	if ucb.err != nil {
		return nil, ucb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ucb.builders))
	nodes := make([]*User, len(ucb.builders))
	mutators := make([]Mutator, len(ucb.builders))
	for i := range ucb.builders {
		func(i int, root context.Context) {
			builder := ucb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*UserMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, ucb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = ucb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ucb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, ucb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ucb *UserCreateBulk) SaveX(ctx context.Context) []*User {
	v, err := ucb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ucb *UserCreateBulk) Exec(ctx context.Context) error {
	_, err := ucb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ucb *UserCreateBulk) ExecX(ctx context.Context) {
	if err := ucb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.User.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.UserUpsert) {
//			SetLogin(v+v).
//		}).
//		Exec(ctx)
func (ucb *UserCreateBulk) OnConflict(opts ...sql.ConflictOption) *UserUpsertBulk {
	ucb.conflict = opts
	return &UserUpsertBulk{
		create: ucb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.User.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ucb *UserCreateBulk) OnConflictColumns(columns ...string) *UserUpsertBulk {
	ucb.conflict = append(ucb.conflict, sql.ConflictColumns(columns...))
	return &UserUpsertBulk{
		create: ucb,
	}
}

// UserUpsertBulk is the builder for "upsert"-ing
// a bulk of User nodes.
type UserUpsertBulk struct {
	create *UserCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.User.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(user.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *UserUpsertBulk) UpdateNewValues() *UserUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(user.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.User.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *UserUpsertBulk) Ignore() *UserUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *UserUpsertBulk) DoNothing() *UserUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the UserCreateBulk.OnConflict
// documentation for more info.
func (u *UserUpsertBulk) Update(set func(*UserUpsert)) *UserUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&UserUpsert{UpdateSet: update})
	}))
	return u
}

// SetLogin sets the "login" field.
func (u *UserUpsertBulk) SetLogin(v string) *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.SetLogin(v)
	})
}

// UpdateLogin sets the "login" field to the value that was provided on create.
func (u *UserUpsertBulk) UpdateLogin() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.UpdateLogin()
	})
}

// SetAge sets the "age" field.
func (u *UserUpsertBulk) SetAge(v int) *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.SetAge(v)
	})
}

// AddAge adds v to the "age" field.
func (u *UserUpsertBulk) AddAge(v int) *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.AddAge(v)
	})
}

// UpdateAge sets the "age" field to the value that was provided on create.
func (u *UserUpsertBulk) UpdateAge() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.UpdateAge()
	})
}

// ClearAge clears the value of the "age" field.
func (u *UserUpsertBulk) ClearAge() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.ClearAge()
	})
}

// SetLocation sets the "location" field.
func (u *UserUpsertBulk) SetLocation(v string) *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.SetLocation(v)
	})
}

// UpdateLocation sets the "location" field to the value that was provided on create.
func (u *UserUpsertBulk) UpdateLocation() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.UpdateLocation()
	})
}

// ClearLocation clears the value of the "location" field.
func (u *UserUpsertBulk) ClearLocation() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.ClearLocation()
	})
}

// SetGender sets the "gender" field.
func (u *UserUpsertBulk) SetGender(v user.Gender) *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.SetGender(v)
	})
}

// UpdateGender sets the "gender" field to the value that was provided on create.
func (u *UserUpsertBulk) UpdateGender() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.UpdateGender()
	})
}

// ClearGender clears the value of the "gender" field.
func (u *UserUpsertBulk) ClearGender() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.ClearGender()
	})
}

// Exec executes the query.
func (u *UserUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the UserCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for UserCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *UserUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
