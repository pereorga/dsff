# DSFF

Codi font de la versió en línia del Diccionari de Sinònims de Frases Fetes ([DSFF](https://dsff.uab.cat/)).

## Dependències

### Opció 1: execució amb Docker

- Docker Compose

### Opció 2: compilació nativa amb Go

- Go 1.24+

### Desenvolupament (opcional)

- Go 1.24+
- Node.js (per a la gestió de dependències i scripts)

Els assets CSS i JavaScript ja estan compilats i comprimits dins del directori `public/`. Si es volguessin recompilar, cal:

- Executar `npm ci` per instal·lar les dependències
- Tenir `brotli` i `zopfli` instal·lats per la compressió dels assets

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
