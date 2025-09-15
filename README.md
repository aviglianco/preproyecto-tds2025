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

1. Generar el parser:
   ```bash
   tree-sitter generate
   ```
2. Instalar las dependencias necesarias (para bindings y tooling):
   ```bash
   npm i
   ```

### Ejecutar el compilador

El programa recibe como entrada un archivo con extensión `.ctds` y genera un archivo de salida con la misma base y extensión `.sint` que contiene el árbol de sintaxis en formato S‑exp.

- Forma recomendada (desde la raíz del proyecto):
  ```bash
  go run . ruta/del/archivo.ctds
  ```

Esto creará `ruta/del/archivo.sint` y mostrará en consola la ruta del archivo generado.

#### Ejemplos

```bash
go run . target_source/preproyecto1.ctds
# -> genera target_source/preproyecto1.sint

go run . target_source/preproyecto2.ctds
# -> genera target_source/preproyecto2.sint
```

#### Compilar binario (opcional)

```bash
go build -o ctds
./ctds path/to/archivo.ctds
```

#### Notas

- Si el archivo de entrada no tiene extensión `.ctds`, el programa fallará con un mensaje de error.
- Cada ejecución sobrescribe el `.sint` si ya existe.

### Ramas correspondientes a cada etapa

Las diferentes etapas del proyecto se desarrollan y entregan de forma progresiva en ramas dedicadas:

- **Scanner y Parser**: `scanner-parser`