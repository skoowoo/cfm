package cfm

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
)

const CTX_ROOT_NAME = "root"

var contexts map[string]*Context

// 'Context' is the config scope for one module
type Context struct {
	name       string
	commands   map[string]*Command
	childs     map[string]*Context
	parent     *Context
	parentName string
	found      bool
	conf       interface{}
}

func NewContext(name string, pName string) *Context {
	ctx := new(Context)
	ctx.name = name
	ctx.commands = make(map[string]*Command)
	ctx.childs = make(map[string]*Context)
	ctx.parentName = pName

	if contexts == nil {
		contexts = make(map[string]*Context)
	}
	if _, ok := contexts[name]; ok {
		return nil
	}
	contexts[name] = ctx

	return ctx
}

func NewRootContext() (root *Context) {
	root = NewContext(CTX_ROOT_NAME, "")
	return
}

func (c *Context) AddCommand(cmd []Command) error {
	for i := 0; i < len(cmd); i++ {
		v := &cmd[i]
		if _, ok := c.commands[v.Name]; ok {
			return errors.New("duplicate command: " + v.Name)
		}
		c.commands[v.Name] = v
	}
	return nil
}

func (c *Context) AddConf(conf interface{}) {
	c.conf = conf
}

func (c *Context) LookupAncestor(name string) *Context {
	for a := c.parent; a != nil; a = a.parent {
		if a.name == name {
			return a
		}
	}
	return nil
}

func (c *Context) Conf() interface{} {
	return c.conf
}

func (c *Context) Found() bool {
	return c.found
}

type CommandSetter func(conf interface{}, field string, args []string) error

type Command struct {
	Name    string
	Field   string
	Default interface{}
	Setter  CommandSetter
}

// set integer value
func CommandSetInt(conf interface{}, field string, args []string) error {
	v, err := GetStructField(conf, field, reflect.Int)
	if err != nil {
		return err
	}

	if val, err := strconv.Atoi(args[0]); err != nil {
		return err
	} else {
		v.SetInt(int64(val))
	}

	return nil
}

// set string value
func CommandSetString(conf interface{}, field string, args []string) error {
	v, err := GetStructField(conf, field, reflect.String)
	if err != nil {
		return err
	}

	val := trimString(args[0])
	v.SetString(val)
	return nil
}

// set on/off value, bool value
func CommandSetOnOff(conf interface{}, field string, args []string) error {
	v, err := GetStructField(conf, field, reflect.Bool)
	if err != nil {
		return err
	}

	switch {
	case args[0] == "on":
		v.SetBool(true)

	case args[0] == "off":
		v.SetBool(false)

	default:
		return errors.New("Not on/off")
	}

	return nil
}

func CommandSetIntArray(conf interface{}, field string, args []string) error {
	v, err := GetStructField(conf, field, reflect.Slice)
	if err != nil {
		return err
	}

	var tmp int
	intType := reflect.TypeOf(tmp)

	l := len(args)
	slice := reflect.MakeSlice(reflect.SliceOf(intType), l, l)

	for i := 0; i < l; i++ {
		val := slice.Index(i)
		if a, err := strconv.Atoi(args[i]); err != nil {
			return err
		} else {
			val.SetInt(int64(a))
		}
	}

	v.Set(slice)
	return nil
}

func CommandSetStringArray(conf interface{}, field string, args []string) error {
	v, err := GetStructField(conf, field, reflect.Slice)
	if err != nil {
		return err
	}

	var tmp string
	strType := reflect.TypeOf(tmp)

	l := len(args)
	slice := reflect.MakeSlice(reflect.SliceOf(strType), l, l)

	for i := 0; i < l; i++ {
		val := slice.Index(i)
		s := trimString(args[i])
		val.SetString(s)
	}

	v.Set(slice)
	return nil
}

type Config struct {
	content     []byte
	path        string
	allContexts map[string]*Context
}

func LoadConfig(path string) *Config {
	c := new(Config)
	c.allContexts = contexts
	c.path = path

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	c.content = content
	return c
}

/*
cmd1 1 2;
cmd2 1 2;

$tcp {
    $http {
        b 1;
    }

    a 1;
}

$udp {
    xx 1;

    $dns {
        c 2;
    }

    d 3;
}
*/
// Call Parse() to parse the config file
func (c *Config) Parse() error {
	if err := c.buildRelation(); err != nil {
		return err
	}
	if err := c.initDefault(); err != nil {
		return err
	}

	ctxStack := newStack()
	rootCtx, ok := c.allContexts[CTX_ROOT_NAME]
	if !ok {
		return errors.New("Not found root context")
	}
	ctxStack.push(rootCtx)

	const (
		swStart = iota
		swComment
		swTryCmd
	)

	var (
		start int
		end   int
		line  int
	)

	line = 0
	state := swStart

	for i := 0; i < len(c.content); i++ {
		ch := c.content[i]

		if ch == '\n' {
			line++
		}

		switch state {
		case swStart:

			if skip(ch) {
				break
			}

			if ch == '}' {
				if ctxStack.size() > 1 {
					ctxStack.pop()
				} else {
					return fmt.Errorf("line %d, unexpected '{'", line)
				}

				break
			}

			if ch == '#' {
				state = swComment
				break
			}

			start = i
			state = swTryCmd
			break

		case swComment:
			if ch == '\n' {
				state = swStart
			}

			break

		case swTryCmd:
			if ch == '{' {
				end = i - 1

				if end <= start {
					return fmt.Errorf("line %d, Invalid context name: %s", line, string(c.content[start:end]))
				}

				name := string(bytes.Trim(c.content[start:end], blankCutSet))

				if ctx, ok := c.allContexts[name]; !ok {
					return fmt.Errorf("line %d, Not found context: %s", line, name)
				} else {
					ctxStack.push(ctx)
					ctx.found = true
				}

				state = swStart
			}

			if skip(ch) {
				break
			}

			if ch == ';' {
				end = i

				ctx := ctxStack.top()
				if err := tryParseCommand(ctx, c.content[start:end]); err != nil {
					return fmt.Errorf("line %d, %s", line, err.Error())
				}

				state = swStart
			}

			break
		}
	}

	if ctxStack.size() != 1 || (state != swStart && state != swComment) {
		return errors.New("config not completion")
	}

	return nil
}

func (c *Config) buildRelation() error {
	for name, ctx := range c.allContexts {
		if ctx.parentName == "" && name == CTX_ROOT_NAME {
			continue
		}

		if parent, ok := c.allContexts[ctx.parentName]; !ok {
			return errors.New("Not found context: " + ctx.parentName)
		} else {
			ctx.parent = parent

			if _, ok := parent.childs[name]; ok {
				errors.New("duplicate context: " + name)
			}
			parent.childs[name] = ctx
		}
	}
	return nil
}

func (c *Config) initDefault() error {
	var dltVal string

	for _, ctx := range c.allContexts {
		for _, cmd := range ctx.commands {

			dlfType := reflect.TypeOf(cmd.Default)
			if dlfType.Kind() != reflect.Int && dlfType.Kind() != reflect.String {
				return errors.New("default value's type error (not int or string)")
			}

			if dlfType.Kind() == reflect.Int {
				v := cmd.Default.(int)
				dltVal = strconv.Itoa(v)
			} else {
				dltVal = cmd.Default.(string)
			}

			args := splitCommandEntry([]byte(dltVal))
			if len(args) == 0 {
				continue
			}

			if err := cmd.Setter(ctx.Conf(), cmd.Field, args); err != nil {
				return err
			}
		}
	}
	return nil
}

func tryParseCommand(ctx *Context, s []byte) error {
	fields := splitCommandEntry(s)
	if len(fields) == 0 {
		return nil
	}

	name := fields[0]
	args := fields[1:]

	cmd, ok := ctx.commands[name]
	if !ok {
		return errors.New("Not found command: " + name)
	}

	if len(args) == 0 {
		return fmt.Errorf("Command \"%s\" not value", name)
	}

	if err := cmd.Setter(ctx.Conf(), cmd.Field, args); err != nil {
		return err
	}

	return nil
}

func init() {
	if contexts == nil {
		contexts = make(map[string]*Context)
	}
}
