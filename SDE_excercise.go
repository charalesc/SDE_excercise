package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var TRACE_LEVEL int

type destinations_struct struct {
	street          string
	interior_number string
	exterior_number string
}

type drivers_struct struct {
	name    string
	company string
}

type data_struct_struct struct {
	drivers_cardinality      int
	destinations_cardinality int
	proposal_list            []proposal_list_struct
	proposal_max             float64
	final_list               []proposal_list_struct
	max_ss                   float64
	free_destinations        []int
	scores                   [][]float64
}

type proposal_list_struct struct {
	driver_index      int
	destination_index int
}

func main() {

	var parameters_quantity int
	var dato string
	var follow_id string
	var err error
	var destinations_file_name string
	var drivers_file_name string
	var number_of_destinations int
	var number_of_drivers int
	var destinations []destinations_struct
	var drivers []drivers_struct
	var scores [][]float64
	var definitive_list []proposal_list_struct
	var max_ss float64

	TRACE_LEVEL = 0
	parameters_quantity = 0

	program_name := path.Base(os.Args[0]) // Name of this program
	follow_id = program_name

	/*

	   Busco los parametros en linea de comandos

	*/

	for index, _ := range os.Args {

		var parametro string
		where_is_found := strings.Index(os.Args[index], "=")

		if where_is_found != -1 {
			parametro = os.Args[index][:where_is_found]
			if where_is_found < len(os.Args[index]) {
				dato = os.Args[index][where_is_found+1:]
			} else {
				dato = ""
			}
		} else {
			parametro = os.Args[index]
		}

		if (parametro == "?") || (strings.ToLower(parametro) == "help") { // OJO verificar el metodo correcto
			PrintLog(follow_id, 0, "It's used by writing:")
			PrintLog(follow_id, 0, "	", os.Args[0], " PAR1=AAAA PAR2=BBBB ...")
			PrintLog(follow_id, 0, "	", os.Args[0], " ARCHIVO_PDF=xxx TRACE_LEVEL=n")
			PrintLog(follow_id, 0, "	DESTINATIONS_FILE File name with the destinations")
			PrintLog(follow_id, 0, "	DRIVERS_FILE File name with de drivers")
			PrintLog(follow_id, 0, "	If not specified the trace level, by default it's 0")
			PrintLog(follow_id, 0, "	Example.- write ", os.Args[0], " it's the same to write ", os.Args[0], " TRACE_LEVEL=0'")
			break
		}

		if parametro == "TRACE_LEVEL" {
			var valor int64

			valor, err = strconv.ParseInt(dato, 10, 0)

			if err == nil {
				TRACE_LEVEL = int(valor)

				if TRACE_LEVEL < 0 {
					TRACE_LEVEL = int(0)
				}
				if TRACE_LEVEL > 30 {
					TRACE_LEVEL = int(30)
				}

				if TRACE_LEVEL > 0 {
					PrintLog(follow_id, 5, "TRACE_LEVEL >", TRACE_LEVEL, "<")
				}
			}
		}

		if parametro == "DESTINATIONS_FILE" {
			destinations_file_name = dato
			parameters_quantity++
		}

		if parametro == "DRIVERS_FILE" {
			drivers_file_name = dato
			parameters_quantity++
		}

		if err != nil {
			break
		}
	}

	if err == nil {
		if parameters_quantity != 2 {
			err = errors.New("Missing parameters")
			PrintLog(follow_id, 0, "It's used by writing:")
			PrintLog(follow_id, 0, "	", os.Args[0], " PAR1=AAAA PAR2=BBBB ...")
			PrintLog(follow_id, 0, "	", os.Args[0], " ARCHIVO_PDF=xxx TRACE_LEVEL=n")
			PrintLog(follow_id, 0, "	DESTINATIONS_FILE File name with the destinations")
			PrintLog(follow_id, 0, "	DRIVERS_FILE File name with de drivers")
			PrintLog(follow_id, 0, "	If not specified the trace level, by default it's 0")
			PrintLog(follow_id, 0, "	Example.- write ", os.Args[0], " it's the same to write ", os.Args[0], " TRACE_LEVEL=0'")
		}
	}

	if err == nil {
		number_of_destinations, destinations, err = ReadDestinationsFile(follow_id, destinations_file_name)
	}

	if err == nil {
		number_of_drivers, drivers, err = ReadDriversFile(follow_id, drivers_file_name)
	}

	if number_of_destinations == 0 || number_of_drivers == 0 {
		if number_of_destinations == 0 {
			err = errors.New("Destinations file is empty")
		}
		if number_of_drivers == 0 {
			err = errors.New("Drivers file is empty")
		}
		if number_of_destinations == 0 && number_of_drivers == 0 {
			err = errors.New("Both files are empty")
		}
	}

	if err == nil {
		scores = SuitabilityScore(follow_id, number_of_destinations, destinations, number_of_drivers, drivers)
		PrintLog(follow_id, 29, program_name, scores)
	}

	var data_struct data_struct_struct

	free_destinations := make([]int, number_of_destinations)
	for index := 0; index < number_of_destinations; index++ {
		free_destinations[index] = index
	}

	final_list := make([]proposal_list_struct, number_of_drivers)

	if number_of_destinations > number_of_drivers {
		PrintLog(follow_id, 0, "  Warning there are not enough drivers")
	}

	data_struct.drivers_cardinality = number_of_drivers
	data_struct.destinations_cardinality = number_of_destinations
	data_struct.free_destinations = free_destinations
	data_struct.scores = scores
	data_struct.final_list = final_list

	definitive_list, max_ss = CalculateMaxSS(follow_id, &data_struct)

	for index := 0; index < len(definitive_list); index++ {
		text := "Send driver " + drivers[definitive_list[index].driver_index].name + " to " + destinations[definitive_list[index].destination_index].street
		fmt.Println(text)
	}
	text := "Final Suitability Score " + fmt.Sprintf("%.2f", max_ss)
	fmt.Println(text)

	if err != nil {
		PrintLog(follow_id, 0, "  Error >", err.Error(), "<")
		PrintLog(follow_id, 0, "Exiting ", program_name)
	} else {
		PrintLog(follow_id, 13, "Exiting ", program_name)
	}

}

/*

 Create the list and calculate the best path to assign drivers to destinations

	Input variables:
	  follow_id string, id to follow the execution on one thread
	  data_struct *data_struct_struct, all the information to calculate

	Output variables:
	  final_list []proposal_list_struct, final list with the best pairs drivers and destinations, based on THE top secret algorithm
			max_ss float64, the total ss of the final path

	Cooments:

*/
func CalculateMaxSS(follow_id string, data_struct *data_struct_struct) (final_list []proposal_list_struct, max_ss float64) {

	PrintLog(follow_id, 23, "Entering CalculateMaxSS")
	PrintLog(follow_id, 29, "  input data data_struct")

	for driver_index := 0; driver_index < data_struct.drivers_cardinality; driver_index++ { // Traverse drivers list

		free_destinations := make([]int, data_struct.destinations_cardinality)
		for index := 0; index < data_struct.destinations_cardinality; index++ {
			free_destinations[index] = index
		}
		data_struct.free_destinations = free_destinations

		CalculateTreeSS(follow_id, driver_index, data_struct)
	}

	final_list = make([]proposal_list_struct, len(data_struct.final_list))
	for index := 0; index < len(data_struct.final_list); index++ {
		final_list[index] = data_struct.final_list[index]
	}
	max_ss = data_struct.max_ss

	PrintLog(follow_id, 29, "  Exit data final_list >", data_struct.final_list, "< max_ss >", data_struct.max_ss, "<")
	PrintLog(follow_id, 23, "Exiting CalculateMaxSS")

	return final_list, max_ss
}

/*

 Recusively will calculate all the paths to obtain the highest ss

	Input variables:
	  follow_id string, id to follow the execution on one thread
	  driver_index int, the number of the driver
	  data_struct *data_struct_struct, all the information to calculate

	Output variables:

	Cooments:

*/
func CalculateTreeSS(follow_id string, driver_index int, data_struct *data_struct_struct) {

	var proposal_list proposal_list_struct
	var temp_number float64
	var input_free_destinations []int

	PrintLog(follow_id, 23, "Entering CalculateTreeSS")
	PrintLog(follow_id, 29, "  input data driver_index >", driver_index, "< data_struct.free_destinations >", data_struct.free_destinations, "<")

	// For backup of the free destinations list
	input_free_destinations = make([]int, len(data_struct.free_destinations))
	for index := 0; index < len(data_struct.free_destinations); index++ {
		input_free_destinations[index] = data_struct.free_destinations[index]
	}

	PrintLog(follow_id, 29, "  CalculateTreeSS driver_index >", driver_index, "< input_free_destinations >", input_free_destinations, "<")

	for destination_index := 0; destination_index < len(input_free_destinations); destination_index++ {

		PrintLog(follow_id, 29, "  CalculateTreeSS calculation with driver_index >", driver_index, "< destination_index >", input_free_destinations[destination_index], "<")

		// Add to the proposal list the ss of this combination
		proposal_list.driver_index = driver_index
		proposal_list.destination_index = input_free_destinations[destination_index] // To do the original free destination list
		temp_number = data_struct.scores[driver_index][input_free_destinations[destination_index]]
		data_struct.proposal_max += temp_number
		data_struct.proposal_list = append(data_struct.proposal_list, proposal_list)

		// Recreate the free destinations list, for use with recursion
		data_struct.free_destinations = nil // Delete all destinations
		for index := 0; index < len(input_free_destinations); index++ {
			if index != destination_index { // Delete the destination used on loop
				data_struct.free_destinations = append(data_struct.free_destinations, input_free_destinations[index]) // Add destination to list
			}
		}

		PrintLog(follow_id, 29, "  CalculateTreeSS recreated driver_index >", driver_index, "< free_destinations >", data_struct.free_destinations, "<")

		if (driver_index + 1) < data_struct.drivers_cardinality { // Go down for the rest of the drivers
			CalculateTreeSS(follow_id, driver_index+1, data_struct)
		} else { // It was the last driver
			PrintLog(follow_id, 29, "  CalculateTreeSS driver_index >", driver_index, "< proposal_max >", data_struct.proposal_max, "< max_ss >", data_struct.max_ss, "<")
			if data_struct.proposal_max > data_struct.max_ss {
				data_struct.max_ss = data_struct.proposal_max
				for driver_number := 0; driver_number < data_struct.drivers_cardinality; driver_number++ {
					data_struct.final_list[driver_number].driver_index = data_struct.proposal_list[driver_number].driver_index
					data_struct.final_list[driver_number].destination_index = data_struct.proposal_list[driver_number].destination_index
				}
				PrintLog(follow_id, 29, "  CalculateTreeSS driver_index >", driver_index, "< final_list >", data_struct.final_list, "< max_ss >", data_struct.max_ss, "<")
			}
		}

		data_struct.proposal_list = data_struct.proposal_list[:len(data_struct.proposal_list)-1] // Remove the last proposal
		data_struct.proposal_max -= temp_number

	}

	PrintLog(follow_id, 29, "  Exit data driver_index >", driver_index, "< final_list >", data_struct.final_list, "< max_ss >", data_struct.max_ss, "<")
	PrintLog(follow_id, 23, "Exiting CalculateTreeSS")
}

/*

 Return an array of suitability score calculated for all the destinations and drivers

	Input variables:
	  follow_id string, id to follow the execution on one thread
	  destination []destinations_struct, array of struct that contains the data of all destinations
	  driver []drivers_struct, array of struct that contains the data of all the drivers

	Output variables:
	  score []float64, array with the suitability score calculated for all the combinantions of destinations and drivers
			err eror, returns the error in case of happening

	Cooments:
	  It creates a matrix with all the calculated suitability score between destinations and drivers

*/
func SuitabilityScore(follow_id string, number_of_destinations int, destinations []destinations_struct, number_of_drivers int, drivers []drivers_struct) (scores [][]float64) {

	var index_destinations int
	var index_drivers int

	PrintLog(follow_id, 13, "Entering SuitabilityScore")
	PrintLog(follow_id, 20, "  input data number_of_destinations >", number_of_destinations, "< destinations >", destinations, "< number_of_drivers >", number_of_drivers, "< drivers >", drivers, "<")

	scores = make([][]float64, number_of_drivers)
	for index := range scores {
		scores[index] = make([]float64, number_of_destinations)
	}

	for index_drivers = 0; index_drivers < number_of_drivers; index_drivers++ {
		for index_destinations = 0; index_destinations < number_of_destinations; index_destinations++ {
			ss := CalculateSuitabilityScore(follow_id, destinations[index_destinations], drivers[index_drivers])
			scores[index_drivers][index_destinations] = ss
		}
	}

	PrintLog(follow_id, 20, "  Exit data score >", scores, "<")
	PrintLog(follow_id, 13, "Exiting SuitabilityScore")

	return scores
}

/*

 Return the suitability score calculated

	Input variables:
	  follow_id string, id to follow the execution on one thread
	  destination destinations_struct, this struct contains all the data of a destination
	  driver drivers_struct, this struct contains all the data of a driver

	Output variables:
	  score float64, the number calculated

	Cooments:
	  It works with a top-secret algorith consider this function for your eyes only

*/
func CalculateSuitabilityScore(follow_id string, destination destinations_struct, driver drivers_struct) (score float64) {

	PrintLog(follow_id, 23, "Entering CalculateSuitabilityScore")
	PrintLog(follow_id, 29, "  input data destination >", destination, "< driver >", driver, "<")

	if Even(len(destination.street)) {
		/*
		   If the length of the shipment's destination street name is even, the base suitability score (SS) is the number of vowels in the driver’s name multiplied by 1.5.
		*/
		vowels_quantity := GetVowelsQuantity(follow_id, driver.name)
		score = float64(vowels_quantity) * 1.5
	} else {
		/*
		   If the length of the shipment's destination street name is odd, the base SS is the number of consonants in the driver’s name multiplied by 1.
		*/
		consonants_quantity := GetConsonantsQuantity(follow_id, driver.name)
		score = float64(consonants_quantity) * 1
	}

	if SearchCommonFactors(follow_id, destination.street, driver.name) {
		/*
		   If the length of the shipment's destination street name shares any common factors (besides 1) with the length of the driver’s name, the SS is increased by 50% above the base SS.
		*/
		score *= 1.5
	}

	PrintLog(follow_id, 29, "  Exit data score >", score, "<")
	PrintLog(follow_id, 23, "Exiting CalculateSuitabilityScore")

	return score
}

/*

 Function to search for commnon mathematical factors on to strings

	Input variables:
	  follow_id string, id to follow the execution on one thread
	  destination_street_name string, string with the street name of destination
			driver_name strings, string with the name of the driver

	Output variables:
	  common_factor bool, variable with the resul of the search, true if the two names has at least one common factor besides 1

	Cooments:
	  The funcition obtain the length of the two names to work
	  The strings could contain characters besides letters
	  Return true if it find at least one common mathematical factor
			It doesn't consider the factor 1

*/
func SearchCommonFactors(follow_id string, destination_street_name string, driver_name string) (common_factor bool) {

	PrintLog(follow_id, 23, "Entering SearchCommonFactors")
	PrintLog(follow_id, 29, "  input data destination_street_name >", destination_street_name, "< driver_name >", driver_name, "<")

	common_factor = false

	destination_factors := Factors(follow_id, len(destination_street_name))
	driver_factors := Factors(follow_id, len(driver_name))

	for index_destination := 0; index_destination < len(destination_factors); index_destination++ {
		for driver_index := 0; driver_index < len(driver_factors); driver_index++ {
			if destination_factors[index_destination] == driver_factors[driver_index] {
				common_factor = true
				break
			}
		}
	}

	PrintLog(follow_id, 29, "  Exit data common_factor >", common_factor, "<")
	PrintLog(follow_id, 23, "Exiting SearchCommonFactors")

	return common_factor
}

/*

 Return the mathematical factors for the number received

	Input variables:
	  follow_id string, id to follow the execution on one thread
	  number int, the number to search for it's mathematical factors

	Output variables:
	  list []int, array list with the mathematical factors (if any) that the number has

	Cooments:
	  The numbers returned will not include the 1
	  List could be empty if the number has no mathematical factors

*/
func Factors(follow_id string, number int) (list []int) {

	PrintLog(follow_id, 23, "Entering Factors")
	PrintLog(follow_id, 29, "  input data number >", number, "<")

	if number > 1 {
		for index := 2; index <= number; index++ {
			if number%index == 0 {
				list = append(list, index)
			}
		}
	}

	PrintLog(follow_id, 29, "  Exit data list >", list, "<")
	PrintLog(follow_id, 23, "Exiting Factors")

	return list
}

/*

 Return the number of vowels on a text

	Input variables:
	  follow_id string, id to follow the execution on one thread
	  input_text string, the text to search for vowels

	Output variables:
	  quantity int, the quantity of vowels in the text

	Cooments:
	  If the text contains spaces or other symbols besides vowels, those will not be counted

*/
func GetVowelsQuantity(follow_id string, input_text string) (quantity int) {

	PrintLog(follow_id, 23, "Entering GetVowelsQuantity")
	PrintLog(follow_id, 29, "  input data input_text >", input_text, "<")

	quantity = 0

	for index := 0; index < len(input_text); index++ {
		character := string(input_text[index])
		switch strings.ToUpper(character) {
		case "A":
			quantity++
		case "E":
			quantity++
		case "I":
			quantity++
		case "O":
			quantity++
		case "U":
			quantity++
		default:
		}
	}

	PrintLog(follow_id, 29, "  Exit data quantity >", quantity, "<")
	PrintLog(follow_id, 23, "Exiting GetVowelsQuantity")

	return quantity
}

/*

 Return the number of consonants on a text

	Input variables:
	  follow_id string, id to follow the execution on one thread
	  input_text string, the text to search for consonants

	Output variables:
	  quantity int, the quantity of consonants in the text

	Cooments:
	  If the text contains spaces or other symbols besides consonants, those will not be counted

*/
func GetConsonantsQuantity(follow_id string, input_text string) (quantity int) {

	PrintLog(follow_id, 23, "Entering GetConsonantsQuantity")
	PrintLog(follow_id, 29, "  input data input_text >", input_text, "<")

	quantity = 0
	base_text := "BCDFGHJKLMNPQRSTVWXYZ"

	for index := 0; index < len(input_text); index++ {
		character := strings.ToUpper(string(input_text[index]))
		where_is_found := strings.Index(base_text, character)
		if where_is_found != -1 { // Found consonant
			quantity++
		}
	}

	PrintLog(follow_id, 29, "  Exit data quantity >", quantity, "<")
	PrintLog(follow_id, 23, "Exiting GetConsonantsQuantity")

	return quantity
}

/*

 Return true if a number is Even, false otherwise

	Input variables:
	  number int, the number to check

	Output variables:
	  none

*/
func Even(number int) bool {
	return number%2 == 0
}

/*

 Return true if a number is Odd, false otherwise

	Input variables:
	  number int, the number to check

	Output variables:
	  none

*/
func Odd(number int) bool {
	return !Even(number)
}

/*

 Show in a especific format the output log

	Input variables:
		 follow_id string, the id that identifies the thread on a parallel procesing
			level int, the level to decide if the text will be printed or not
			text ...interface{}, the interface to be printed

	Output variables:
	  none

	Coments:
	 Check that the trace level be great than or equal of the input parameter
  The printed format is as follow: follow_id + date_hour + input interface

*/
func PrintLog(follow_id string, level int, text ...interface{}) {

	var output_string string
	const layout = "20060102150405.000"
	var date_text string

	date_text = time.Now().Format(layout)

	if level <= TRACE_LEVEL {

		output_string = follow_id + " " + date_text + " "

		fmt.Print(output_string)
		fmt.Println(text)
	}
}

/*

 Read destinations file into it's corresponding list of destinations

	Input variables:
	  follow_id string, id to follow the execution on one thread
		 file_name string, The name of the file to be read and parsed into lines

	Output variables:
	  lines []string, text array begining in cero
			err error, if filled contains the text of the error

	Coments:
			If the input text, doesn't have text, doesn't return any destination
			Each destination on a diferent line
			The file structure to read is as follows:
	    street;interior_number;exterior_number

*/
func ReadDestinationsFile(follow_id string, file_name string) (number_of_destinations int, destinations []destinations_struct, err error) {

	var all_text string
	var line_quantity int64
	var lines []string
	var destination destinations_struct

	PrintLog(follow_id, 13, "Entering ReadDestinationsFile")
	PrintLog(follow_id, 20, "  input data file_name >", file_name, "<")

	number_of_destinations = 0

	file_text, err := ioutil.ReadFile(file_name)
	if err != nil {
		temp_text := "File does not exist " + file_name
		err = errors.New(temp_text)
	} else {
		all_text = string(file_text)
	}

	if err == nil {
		line_quantity, lines = DivideByLines(all_text)

		if line_quantity == 0 {
			err = errors.New("Destinations Empty File")
		}
	}

	if err == nil {
		for index := 0; index < int(line_quantity); index++ {
			word_quantity, words := DivideBySymbol(lines[index], ";")

			PrintLog(follow_id, 29, "  word_quantity >", word_quantity, "< words >", words, "<")

			if word_quantity >= 1 {
				destination.street = words[0]
				if word_quantity >= 2 {
					destination.interior_number = words[1]
					if word_quantity >= 3 {
						destination.exterior_number = words[2]
					}
				}

				destinations = append(destinations, destination)
				number_of_destinations++
			}
		}
	}

	if err != nil {
		PrintLog(follow_id, 0, "  Error >", err.Error(), "<")
		PrintLog(follow_id, 0, "Exiting ReadDestinationsFile ")
	} else {
		PrintLog(follow_id, 20, "  Exit data number_of_destinations >", number_of_destinations, "< destinations >", destinations, "<")
		PrintLog(follow_id, 13, "Exiting ReadDestinationsFile")
	}

	return number_of_destinations, destinations, err
}

/*

 Read drivers file into it's corresponding list of drivers

	Input variables:
	  follow_id string, id to follow the execution on one thread
		 file_name string, The name of the file to be read and parsed into lines

	Output variables:
	  drivers []string, text array begining in cero
			err error, if filled contains the text of the error

	Coments:
			If the input text, doesn't have text, doesn't return any driver
			Each driver on a diferent line
			The file structure to read is as follows:
	    name;company

*/
func ReadDriversFile(follow_id string, file_name string) (number_of_drivers int, drivers []drivers_struct, err error) {

	var all_text string
	var line_quantity int64
	var lines []string
	var driver drivers_struct

	PrintLog(follow_id, 13, "Entering ReadDriversFile")
	PrintLog(follow_id, 20, "  input data file_name >", file_name, "<")

	number_of_drivers = 0

	file_text, err := ioutil.ReadFile(file_name)
	if err != nil {
		temp_text := "File does not exist " + file_name
		err = errors.New(temp_text)
	} else {
		all_text = string(file_text)
	}

	if err == nil {
		line_quantity, lines = DivideByLines(all_text)

		if line_quantity == 0 {
			err = errors.New("Drivers Empty File")
		}
	}

	if err == nil {
		for index := 0; index < int(line_quantity); index++ {
			word_quantity, words := DivideBySymbol(lines[index], ";")

			PrintLog(follow_id, 29, "  word_quantity >", word_quantity, "< words >", words, "<")

			if word_quantity >= 1 {
				driver.name = words[0]
				if word_quantity >= 2 {
					driver.company = words[1]
				}

				drivers = append(drivers, driver)
				number_of_drivers++
			}
		}
	}

	if err != nil {
		PrintLog(follow_id, 0, "  Error >", err.Error(), "<")
		PrintLog(follow_id, 0, "Leaving ReadDriversFile ")
	} else {
		PrintLog(follow_id, 20, "  Exit data number_of_drivers >", number_of_drivers, "< drivers >", drivers, "<")
		PrintLog(follow_id, 13, "Leaving ReadDriversFile")
	}

	return number_of_drivers, drivers, err
}

/*

 Separate an input text on indivdual lines

	Input variables:
	  input_text strig, contains the input text to be separated

	Output variables:
	  quantity int64, nuber of words found and returned
	  lines []string, text array begining in cero

	Coments:
			If the input text, doesn't have text, doesn't return any line
			The separator can be CR, LF or any combination
			It gives preference to LF before CR, if the input text has only LF, it use it as separator and remove the CR

*/
func DivideByLines(input_text string) (quantity int64, lines []string) {

	var index int
	var where_is_found int

	quantity = 0

	for index = 0; index < len(input_text); index++ {
		where_is_found = strings.Index(input_text[index:], "\n")
		if where_is_found != -1 { // Search if contains LF
			input_text = strings.Replace(input_text, "\r", "", -1) // Remove any CR it can contains
			lines = append(lines, input_text[index:index+where_is_found])
			index += where_is_found
			quantity++
		} else {
			if strings.Index(input_text, "\r") != -1 { // Search only for CR
				lines = append(lines, input_text[index:index+where_is_found])
				index += where_is_found
				quantity++
			} else { // Doesn't has line separator or the last line dosen't have one
				lines = append(lines, input_text[index:])
				quantity++
				break
			}
		}
	}

	return quantity, lines
}

/*

 Separate an input text on indivdual word by a symbol

	Input variables:
	  input_text strig, contains the input text to be separated
			symbol string, contains the text symbol to use as separator

	Output variables:
	  quantity int64, number of words found and returned
	  words []string, text array begining in cero

	Coments:
			If the input text, doesn't have text, dones't return any word

*/
func DivideBySymbol(input_text string, symbol string) (quantity int64, words []string) {

	var original_size int
	var index int
	var cut int
	var exit bool
	var subchain string
	var on_word bool

	original_size = len(input_text)
	quantity = 0
	cut = 0
	on_word = false

	if original_size > 0 {
		exit = false
		for exit == false {
			if cut < original_size {
				subchain = input_text[cut:original_size]
				on_word = false
			} else {
				exit = true
				break
			}

			if exit == false {
				for index = 0; index < len(subchain); index++ {
					if subchain[index:index+1] == symbol { // Here we cut
						if on_word == true {
							words = append(words, subchain[0:index])
							quantity++
							cut += (index + 1)
							break
						}
					} else {
						if on_word == false {
							on_word = true
							cut += index
							subchain = input_text[cut:original_size]
						}
					}
				}

				if index == len(subchain) { // If it doesn't find the symbol, return the subchain as word
					cut += index
					if on_word == true {
						words = append(words, subchain)
						quantity++
						exit = true
					}
				}
			}
		}
	}

	return quantity, words
}
