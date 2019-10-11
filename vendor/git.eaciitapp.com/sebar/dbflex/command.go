package dbflex

import "github.com/eaciit/toolkit"

// ICommand is interface abstraction fo all command should be supported by each driver
type ICommand interface {
	Reset() ICommand
	Select(...string) ICommand
	From(string) ICommand
	Where(*Filter) ICommand
	OrderBy(...string) ICommand
	GroupBy(...string) ICommand

	Aggr(...*AggrItem) ICommand
	Insert(...string) ICommand
	Update(...string) ICommand
	Delete() ICommand
	Save() ICommand

	Take(int) ICommand
	Skip(int) ICommand

	Command(interface{}) ICommand
	SQL(string) ICommand
}

// CommandBase is base struct for any struct that implement ICommand for ease of implementation
type CommandBase struct {
	items []*QueryItem
}

// Reset base implementation of Reset method
func (b *CommandBase) Reset() ICommand {
	b.items = []*QueryItem{}
	return b
}

// Select base implementation of Select method
func (b *CommandBase) Select(fields ...string) ICommand {
	b.items = append(b.items, &QueryItem{QuerySelect, fields})
	return b
}

// From base implementation of From method
func (b *CommandBase) From(name string) ICommand {
	b.items = append(b.items, &QueryItem{QueryFrom, name})
	return b
}

// Where base implementation of Where method
func (b *CommandBase) Where(f *Filter) ICommand {
	b.items = append(b.items, &QueryItem{QueryWhere, f})
	return b
}

// OrderBy base implementation of OrderBy method
func (b *CommandBase) OrderBy(fields ...string) ICommand {
	b.items = append(b.items, &QueryItem{QueryOrder, fields})
	return b
}

// GroupBy base implementation of GroupBy method
func (b *CommandBase) GroupBy(fields ...string) ICommand {
	b.items = append(b.items, &QueryItem{QueryGroup, fields})
	return b
}

// Aggr base implementation of Aggr method
func (b *CommandBase) Aggr(aggritems ...*AggrItem) ICommand {
	b.items = append(b.items, &QueryItem{QueryAggr, aggritems})
	return b
}

// Insert base implementation of Insert method
func (b *CommandBase) Insert(fields ...string) ICommand {
	b.items = append(b.items, &QueryItem{QueryInsert, fields})
	return b
}

// Update base implementation of Update method
func (b *CommandBase) Update(fields ...string) ICommand {
	b.items = append(b.items, &QueryItem{QueryUpdate, fields})
	return b
}

// Delete base implementation of Delete method
func (b *CommandBase) Delete() ICommand {
	b.items = append(b.items, &QueryItem{QueryDelete, true})
	return b
}

// Save base implementation of Save method
func (b *CommandBase) Save() ICommand {
	b.items = append(b.items, &QueryItem{QuerySave, true})
	return b
}

// Take base implementation of Take method
func (b *CommandBase) Take(n int) ICommand {
	b.items = append(b.items, &QueryItem{QueryTake, n})
	return b
}

// Skip base implementation of Skip method
func (b *CommandBase) Skip(n int) ICommand {
	b.items = append(b.items, &QueryItem{QuerySkip, n})
	return b
}

// Command base implementation of Command method
func (b *CommandBase) Command(command interface{}) ICommand {
	b.items = append(b.items,
		&QueryItem{QueryCommand, toolkit.M{}.Set("command", command)})
	return b
}

// SQL base implementation of SQL method
func (b *CommandBase) SQL(sql string) ICommand {
	b.items = []*QueryItem{&QueryItem{QuerySQL, sql}}
	return b
}
