Solución alternativa al repositorio principal.

Solamente queda adaptar el proceso `abeja.go` para que se adapte.

Consiste en solamente 3 colas en total, y la abeja acaba al recibir un mensaje de bote roto.
Los demás mensajes (para permisos) se consumen solamente por la abeja a la que le es dado el permiso, siendo este un mensaje el cual al abeja pertinente le da un acknowledge manual (en la solución principal es automático).

TODO:
    - Falta mejorar el código
    - `abeja.go` no gestiona bien los mensajes y/o el `oso.go` tampoco ("porción 12" cuando el máximo es 10)

Por lo demás todo bien.
