# config
Simple config information for Go application saved in json format

### Usage
Load from repo
```
go get -u github.com/juragan360/config

```
then import it on any Go app
```
import "github.com/juragan360/config"
```

### Initiate config
Initiate config from specific location
```
if e := config.SetConfigFile("/tmp/app1/config.json"); e!= nil {
	panic(e.Error())
}
```

or 
Initiate config without specify file, then it will create config.json file on same directory of executable
```
if e := config.SetConfigFile(""); e!= nil {
	panic(e.Error())
}
```

### Set a config value
```
config.Set("FullName", "Arief Darmawan")
if e := config.Write(); e != nil {
	panic(e.Error())
}
```

### Get a config value
```
s := config.Get("FullName").(string)
if s != "Arief Darmawan" {
	panic("Unable to read value. Expected 'Arief Darmawan' got '" + s + "'")
}
```
