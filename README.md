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

## 📦 Instalação via GitHub

Para instalar DIRETAMENTE do repositório, sem precisar baixar o código manualmente:

1.  **Instale com Go**:
    ```bash
    go install github.com/dvsalvaya/mangaka/cmd/mangaka@latest
    ```

2.  **Verifique o PATH**:
    Certifique-se de que a pasta de binários do Go está no seu PATH.
    -   *Geralmente*: `%USERPROFILE%\go\bin` (Windows) ou `$HOME/go/bin` (Linux/Mac).

3.  **Use**:
    Agora você pode digitar apenas:
    ```bash
    mangaka
    ```

## 🎮 Como Usar (Código Fonte)

Se preferir rodar localmente para desenvolvimento:

```bash
git clone https://github.com/dvsalvaya/mangaka.git
cd mangaka
go mod tidy
go run cmd/mangaka/main.go
```

## ⚠️ Notas

-   O download dos capítulos é feito para a pasta temporária do sistema e limpo após o uso (exceto o CBZ que é passado pro leitor).
-   A API do MangaDex possui rate limits strict; se falhar, aguarde um pouco.