# PREPROYECTO · TDS 2025

Este repositorio contiene la implementación de la consigna del **preproyecto** de la materia **Taller de Diseño de Software (TDS)**. El trabajo consiste en la construcción de un **pequeño compilador** para un lenguaje sencillo.

## Integrantes

- De Ipola Tobar, Guillermo Javier
- Salas Piñero, Pedro
- Viglianco, Agustín

## Requisitos

- **Go** instalado.
- **tree-sitter-cli** instalado.
- **npm** instalado.

## Ejecución

Por primera vez:

1. Cambiar a la rama `code-generator`:
   ```bash
   $ git checkout code-generator
   ```
2. Ejecutar el generador de parser:
    ```bash
    $ tree-sitter generate
    ```
3. Instalar las dependencias necesarias:
    ```bash
    $ npm i
    ```
4. Ejecutar el compilador desde la raíz del proyecto:
   ```bash
   $ go run .
   ```

Luego de la primera vez, y siempre y cuando no se modifique la gramática, solo es necesario ejecutar el paso 4.