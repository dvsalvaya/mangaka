# Mangaka CLI (MangaDex Edition)

Uma CLI interativa para ler mangás diretamento no terminal, usando a API do **MangaDex** e o leitor **Zathura**.

## 🚀 Funcionalidades

- **Fonte de Dados**: MangaDex (API v5).
- **Leitura**: Baixa capítulos automaticamente e abre no **Zathura** (formato CBZ).
- **Busca & Navegação**: Menu interativo estilo GoAnime.
- **Favoritos**: Gestão local de favoritos.

## 🛠️ Pré-requisitos

1.  **Go 1.20+**
2.  **Zathura**: Deve estar instalado e no PATH do sistema.
    -   *Linux*: `sudo apt install zathura zathura-cb`
    -   *Windows*: Instale via MSYS2 ou WSL, ou certifique-se de que o executável `zathura` está acessível no cmd.

## 🎮 Como Usar

1.  **Rodar**:
    ```bash
    go mod tidy
    go run cmd/mangaka/main.go
    ```

2.  **Navegar**:
    -   `Search Manga` -> Digite o nome -> Selecione.
    -   `List Chapters` -> Escolha o capítulo.
    -   O Mangaka irá baixar as páginas, criar um arquivo `.cbz` temporário e abrir o Zathura.

## ⚠️ Notas

-   O download dos capítulos é feito para a pasta temporária do sistema e limpo após o uso (exceto o CBZ que é passado pro leitor).
-   A API do MangaDex possui rate limits strict; se falhar, aguarde um pouco.