package alias

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Alias struct {
	Handle     string   `json:"handle"`
	Characters []string `json:"characters"`
}

type Aliases struct {
	List []Alias `json:"aliases"`
}

var (
	ActiveAliases Aliases
)

func (aliases Aliases) PrintAliases() error {
	if len(aliases.List) == 0 {
		return fmt.Errorf("PrintAliases(): there are no aliases to print")
	}
	fileName := "alias.json"
	fmt.Println("Print Aliases: " + fileName)
	//Increment checkinCounts
	fmt.Println("Alias Count:", len(aliases.List))
	for _, alias := range aliases.List {
		//Ensure player is on the current player list
		fmt.Printf("Handle(Alias):%s\n", alias.Handle)
		for index, character := range alias.Characters {
			fmt.Printf("%d) %s\n", index+1, character)
		}

	}
	return nil
}

// Adds a character to the specified alias
func (a *Alias) AddCharacter(character string) {
	a.Characters = append(a.Characters, character)
	fmt.Printf("%s added as an alias of handle: %s\n", character, a.Handle)
}

// Removes a character from the specified alias
func (a *Alias) RemoveCharacter(character string) {
	for i, c := range a.Characters {
		if c == character {
			a.Characters = append(a.Characters[:i], a.Characters[i+1:]...)
			return
		}
	}
}

// Checks if a character is part of the specified alias
func (a *Alias) HasCharacter(character string) bool {
	for _, c := range a.Characters {
		if c == character {
			return true
		}
	}
	return false
}

// Adds an alias to the ActiveAliases list
func AddAlias(characterName, handle string) error {
	if IsNameHandle(handle) {
		fmt.Printf("%s is already a handle. Going to add %s to it.\n", handle, characterName)
		selectedAlias, err := GetHandleAlias(handle)
		fmt.Printf("The selected handle is: %s\n", selectedAlias.Handle)
		if err != nil {
			return fmt.Errorf("AddAlias(): GetHandleAlias(): %w", err)
		}
		selectedAlias.AddCharacter(characterName)
	} else {
		fmt.Printf("%s is not currently a handle. Creating new alias for %s, under that new handle.\n", handle, characterName)
		// Create a new alias and add it to the list
		newAlias := Alias{
			Handle:     handle,
			Characters: []string{characterName}}
		ActiveAliases.List = append(ActiveAliases.List, newAlias)
		fmt.Printf("%s added as an alias of handle: %s\n", characterName, handle)
	}
	err := SaveAliases()
	if err != nil {
		return fmt.Errorf("AddAlias(): SaveAliases(): %w", err)
	}
	return nil
}

// Checks if a specified character is present in the alias list
func HasCharacter(character string) bool {
	for _, a := range ActiveAliases.List {
		if a.HasCharacter(character) {
			return true
		}
	}
	return false
}

// Get the handle associated with a characters name
func GetAliasHandle(character string) string {
	for _, a := range ActiveAliases.List {
		if a.HasCharacter(character) {
			return a.Handle
		}
	}
	return ""
}

// Get the alias associated with the provided handle
func GetHandleAlias(handle string) (*Alias, error) {
	for index, a := range ActiveAliases.List {
		if a.Handle == handle {
			return &ActiveAliases.List[index], nil
		}
	}
	return nil, fmt.Errorf("GetAlias(): no alias found for handle: %s", handle)
}

// Check if the provided name is a handle in the active aliases
func IsNameHandle(name string) bool {
	for _, a := range ActiveAliases.List {
		if a.Handle == name {
			return true
		}
	}
	return false
}

// Check if the provided character name has an associated handle in the active aliases
func HasHandle(characterName string) bool {
	for _, a := range ActiveAliases.List {
		if a.HasCharacter(characterName) {
			return true
		}
	}
	return false
}

//Attempts to get the associated alias handle, if none exists, returns the character name
func TryToGetHandle(characterName string) string {
	if IsNameHandle(characterName) {
		return characterName
	} else {
		if HasHandle(characterName) {
			charHandle := GetAliasHandle(characterName)
			return charHandle
		}
	}
	return characterName
}

// Save the alias list to a json file
func SaveAliases() error {
	fmt.Println("Saving to alias file...")
	file, err := json.MarshalIndent(ActiveAliases, "", " ")
	if err != nil {
		return fmt.Errorf("SaveAliases(): failed to marshal ActiveAliases: %w", err)
	}

	err = ioutil.WriteFile("./alias.json", file, 0644)
	if err != nil {
		return fmt.Errorf("SaveAliases(): failed to write to ActiveAliases: %w", err)
	}

	fmt.Println("Alias save successful!")

	return nil

}

func ReadAliases() error {
	fmt.Println("Reading alias file...")
	created, err := checkAliasFile()
	if err != nil {
		return fmt.Errorf("checkAliasFile(): %w", err)
	}
	if created {
		return nil
	}
	file, err := ioutil.ReadFile("./alias.json")
	if err != nil {
		return fmt.Errorf("ReadAliases(): failed to read alias file: %w", err)
	}

	//fmt.Println(string(file))

	err = json.Unmarshal(file, &ActiveAliases)
	if err != nil {
		return fmt.Errorf("ReadAliases(): failed to unmarshal alias file: %w", err)
	}

	if err == nil {
		fmt.Println("Alias load successful!")
	}
	return nil
}

func checkAliasFile() (bool, error) {
	createdFile := false
	// Check if the config.json file exists
	if _, err := os.Stat("./alias.json"); os.IsNotExist(err) {
		fmt.Println("Alias file not found, creating new one...")
		err := SaveAliases()
		if err != nil {
			return false, fmt.Errorf("checkAliasFile(): SaveAliases(): %w", err)
		}
		createdFile = true
	}
	return createdFile, nil
}
