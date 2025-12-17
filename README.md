# DSFF

_Catalan version available below / Versió en català disponible a continuació_

Source code for the online version of the Diccionari de Sinònims de Frases Fetes (Dictionary of Synonyms of Idiomatic Expressions) ([DSFF](https://dsff.uab.cat/)).

## Dependencies

### Option 1: Running with Docker

- Docker Compose

### Option 2: Native Build with Go

- Go 1.24+

### Development (optional)

- Go 1.24+
- Node.js (for dependency management and scripts)

## Building and Running

### Option 1 (Docker)

```
docker compose up
```

### Option 2 (Go)

```
cd go/ && go build -o ../dsff && cd ..
./dsff
```

## Copyright and License

Copyright (c) Pere Orga Esteve <pere@orga.cat>, 2025.

The source code of this project is distributed under the [AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.html) license or later.

## Dictionary Data

The dictionary data is not included in this repository and is licensed separately from the source code. A small sample data file (`data.json.gz`) containing the concept "ALLIBERAR" is included to run the application.

For more details about the dictionary credits, see: [https://dsff.uab.cat/credits](https://dsff.uab.cat/credits).

---

Codi font de la versió en línia del Diccionari de Sinònims de Frases Fetes ([DSFF](https://dsff.uab.cat/)).

## Dependències

### Opció 1: execució amb Docker

- Docker Compose

### Opció 2: compilació nativa amb Go

- Go 1.24+

### Desenvolupament (opcional)

- Go 1.24+
- Node.js (per a la gestió de dependències i scripts)

## Compilació i execució

### Opció 1 (Docker)

```
docker compose up
```

### Opció 2 (Go)

```
cd go/ && go build -o ../dsff && cd ..
./dsff
```

## Copyright i llicència

Copyright (c) Pere Orga Esteve <pere@orga.cat>, 2025.

El codi font d'aquest projecte es distribueix amb la llicència [AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.html)
o superior.

## Dades del diccionari

Les dades del diccionari no s'inclouen en aquest repositori i tenen una llicència diferent de la del codi font.
S'inclou un petit fitxer de dades de mostra (`data.json.gz`) amb el concepte «ALLIBERAR» per a poder executar
l'aplicació.

Per a més detalls sobre els crèdits del diccionari, vegeu: [https://dsff.uab.cat/credits](https://dsff.uab.cat/credits).
