Usage
=====
```
docker-compose up
```

Methods
=====
```
processXlsx(link string, id in) (countAdded, countUpdated, countDeleted, countErr int)

Params:
* link - path to xlsx file
* id - id of seller

Return:
* countAdded - added offers
* countUpdated - updated offers
* countDeleted - deleted offers
* countErr - number of errors accured
```
```
getOffers(mask uint8, params ...interface{}) (names []string, err error)

Params:
* mask - number from 0-7 used to check the passed parameters
(example: 111 - meaning all three are passed, 010 - only second one)
* params - variadic parameters

Return:
* names - offers names
* err - possible error
```
