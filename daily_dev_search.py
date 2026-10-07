#!/usr/bin/env python3
"""
Script para fazer requisições na API pública do Daily.dev
"""

import requests
import json
from getpass import getpass
import urllib3

# Desabilita avisos de SSL para ambientes corporativos
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

BASE_URL = "https://api.daily.dev/public/v1"


def get_bearer_token():
    """Solicita o bearer token ao usuário"""
    print("=== Daily.dev Public API ===")
    token = getpass("Digite seu Bearer Token do Daily.dev: ")
    return token.strip()


def make_request(method, endpoint, token, params=None, data=None):
    """
    Faz uma requisição genérica na API do Daily.dev
    
    Args:
        method: Método HTTP (GET, POST, PUT, PATCH, DELETE)
        endpoint: Endpoint da API
        token: Bearer token de autenticação
        params: Parâmetros de query
        data: Dados para o corpo da requisição
    
    Returns:
        JSON com a resposta da requisição
    """
    url = f"{BASE_URL}{endpoint}"
    
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    try:
        if method.upper() == "GET":
            response = requests.get(url, headers=headers, params=params, verify=False)
        elif method.upper() == "POST":
            response = requests.post(url, headers=headers, params=params, json=data, verify=False)
        elif method.upper() == "PUT":
            response = requests.put(url, headers=headers, params=params, json=data, verify=False)
        elif method.upper() == "PATCH":
            response = requests.patch(url, headers=headers, params=params, json=data, verify=False)
        elif method.upper() == "DELETE":
            response = requests.delete(url, headers=headers, params=params, verify=False)
        else:
            print(f"Método não suportado: {method}")
            return None
        
        response.raise_for_status()
        
        # Retorna None para respostas vazias (DELETE, etc)
        if not response.text:
            return {"status": "success"}
            
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Erro na requisição: {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"Detalhes: {e.response.text}")
        return None


def display_menu():
    """Exibe o menu de opções de requisições"""
    print("\n" + "=" * 60)
    print("Escolha o tipo de requisição:")
    print("=" * 60)
    
    categories = {
        "1": ("Feeds", [
            ("1", "GET /feeds/foryou - Feed personalizado 'For You'"),
            ("2", "GET /feeds/popular - Feed com posts populares/trending"),
            ("3", "GET /feeds/discussed - Feed com posts com discussões"),
            ("4", "GET /feeds/tag/{tag} - Posts por tag"),
            ("5", "GET /feeds/source/{source} - Posts por fonte"),
        ]),
        "2": ("Posts", [
            ("1", "GET /posts/{id} - Detalhes de um post por ID"),
            ("2", "GET /posts/{id}/comments - Comentários de um post"),
        ]),
        "3": ("Search", [
            ("1", "GET /search/posts - Buscar posts por palavra-chave"),
            ("2", "GET /search/tags - Buscar tags por nome"),
            ("3", "GET /search/sources - Buscar fontes/publishers por nome"),
        ]),
        "4": ("Bookmarks", [
            ("1", "GET /bookmarks/ - Obter posts marcados"),
            ("2", "POST /bookmarks/ - Adicionar posts aos marcadores"),
            ("3", "GET /bookmarks/search - Buscar dentro dos marcadores"),
            ("4", "GET /bookmarks/lists - Obter listas de marcadores"),
            ("5", "POST /bookmarks/lists - Criar nova lista de marcadores"),
            ("6", "DELETE /bookmarks/lists/{id} - Deletar lista de marcadores"),
            ("7", "DELETE /bookmarks/{id} - Remover post dos marcadores"),
            ("8", "PATCH /bookmarks/{id} - Mover marcador para lista (Plus)"),
        ]),
        "5": ("Custom Feeds", [
            ("1", "GET /feeds/custom/ - Listar feeds personalizados"),
            ("2", "POST /feeds/custom/ - Criar novo feed personalizado"),
            ("3", "GET /feeds/custom/{feedId} - Obter posts de um feed personalizado"),
            ("4", "PATCH /feeds/custom/{feedId} - Atualizar configurações do feed"),
            ("5", "DELETE /feeds/custom/{feedId} - Deletar feed personalizado"),
            ("6", "GET /feeds/custom/{feedId}/info - Obter metadados do feed"),
            ("7", "PATCH /feeds/custom/{feedId}/advanced - Atualizar configurações avançadas"),
            ("8", "GET /feeds/custom/advanced-settings - Obter configurações avançadas disponíveis"),
        ]),
        "6": ("Feed Filters", [
            ("1", "GET /feeds/filters/ - Obter configurações globais do feed"),
            ("2", "POST /feeds/filters/tags/follow - Seguir tags globalmente"),
            ("3", "POST /feeds/filters/tags/unfollow - Deixar de seguir tags"),
            ("4", "POST /feeds/filters/tags/block - Bloquear tags"),
            ("5", "POST /feeds/filters/tags/unblock - Desbloquear tags"),
            ("6", "POST /feeds/filters/sources/follow - Seguir fontes globalmente"),
            ("7", "POST /feeds/filters/sources/unfollow - Deixar de seguir fontes"),
            ("8", "POST /feeds/filters/sources/block - Bloquear fontes"),
            ("9", "POST /feeds/filters/sources/unblock - Desbloquear fontes"),
        ]),
        "7": ("Notifications", [
            ("1", "GET /notifications/ - Obter notificações do usuário"),
            ("2", "GET /notifications/unread/count - Contar notificações não lidas"),
            ("3", "POST /notifications/read - Marcar todas as notificações como lidas"),
        ]),
        "8": ("Profile", [
            ("1", "GET /profile/ - Obter perfil do usuário atual"),
            ("2", "PATCH /profile/ - Atualizar perfil do usuário"),
        ]),
        "9": ("Stack (Tech Stack)", [
            ("1", "GET /profile/stack/search - Buscar ferramentas/tecnologias por nome"),
            ("2", "GET /profile/stack/ - Obter stack tecnológico do usuário"),
            ("3", "POST /profile/stack/ - Adicionar ferramenta ao stack"),
            ("4", "PATCH /profile/stack/{id} - Atualizar item do stack"),
            ("5", "DELETE /profile/stack/{id} - Remover ferramenta do stack"),
            ("6", "PUT /profile/stack/reorder - Reordenar itens do stack"),
        ]),
        "10": ("Experiences", [
            ("1", "GET /profile/experiences/ - Obter experiências do usuário"),
            ("2", "POST /profile/experiences/ - Criar nova experiência"),
            ("3", "GET /profile/experiences/{id} - Obter experiência específica por ID"),
            ("4", "PUT /profile/experiences/{id} - Atualizar experiência existente"),
            ("5", "DELETE /profile/experiences/{id} - Deletar experiência"),
        ]),
        "11": ("Tags", [
            ("1", "GET /tags/ - Obter todas as tags"),
        ]),
        "12": ("Recommend", [
            ("1", "GET /recommend/keyword - [EXPERIMENTAL] Recomendar artigos por palavra-chave"),
            ("2", "GET /recommend/semantic - [DEPRECATED] Recomendação semântica"),
        ]),
        "13": ("Signup", [
            ("1", "POST /signup/ - Criar conta daily.dev (sem token)"),
        ]),
    }
    
    for cat_num, (cat_name, options) in categories.items():
        print(f"\n{cat_num}. {cat_name}")
        for opt_num, opt_desc in options:
            print(f"   {cat_num}{opt_num}. {opt_desc}")
    
    print("\n0. Sair")
    print("=" * 60)
    
    return categories


def get_endpoint_params(endpoint):
    """
    Solicita parâmetros necessários para o endpoint
    
    Args:
        endpoint: String do endpoint (ex: "/feeds/tag/{tag}")
    
    Returns:
        Dicionário com os parâmetros
    """
    params = {}
    
    # Verifica se há parâmetros de path
    if "{tag}" in endpoint:
        tag = input("Digite o nome da tag: ").strip()
        if tag:
            endpoint = endpoint.replace("{tag}", tag)
    
    if "{source}" in endpoint:
        source = input("Digite o ID ou handle da fonte: ").strip()
        if source:
            endpoint = endpoint.replace("{source}", source)
    
    if "{id}" in endpoint:
        id_val = input("Digite o ID: ").strip()
        if id_val:
            endpoint = endpoint.replace("{id}", id_val)
    
    if "{feedId}" in endpoint:
        feed_id = input("Digite o ID do feed: ").strip()
        if feed_id:
            endpoint = endpoint.replace("{feedId}", feed_id)
    
    return endpoint, params


def get_query_params(endpoint):
    """
    Solicita parâmetros de query comuns
    
    Args:
        endpoint: String do endpoint
    
    Returns:
        Dicionário com os parâmetros de query
    """
    params = {}
    
    # Parâmetros comuns
    if input("Deseja definir limit? (padrão: 20) [s/N]: ").lower() == 's':
        limit = input("Digite o limit (1-50): ").strip()
        if limit:
            params['limit'] = int(limit)
    
    if input("Deseja definir cursor (paginação)? [s/N]: ").lower() == 's':
        cursor = input("Digite o cursor: ").strip()
        if cursor:
            params['cursor'] = cursor
    
    # Parâmetros específicos
    if "search/posts" in endpoint or "recommend" in endpoint:
        query = input("Digite o termo de busca: ").strip()
        if query:
            params['q'] = query
        
        if input("Deseja filtrar por tempo? [s/N]: ").lower() == 's':
            time_filter = input("Digite o filtro (day, week, month, year, all): ").strip()
            if time_filter:
                params['time'] = time_filter
    
    if "search/tags" in endpoint or "search/sources" in endpoint:
        query = input("Digite o termo de busca: ").strip()
        if query:
            params['q'] = query
    
    if "feeds/popular" in endpoint:
        if input("Deseja filtrar por tags? [s/N]: ").lower() == 's':
            tags = input("Digite as tags separadas por vírgula: ").strip()
            if tags:
                params['tags'] = tags
    
    if "feeds/discussed" in endpoint:
        if input("Deseja definir período (dias)? [s/N]: ").lower() == 's':
            period = input("Digite o período (1-30): ").strip()
            if period:
                params['period'] = int(period)
        
        if input("Deseja filtrar por tag? [s/N]: ").lower() == 's':
            tag = input("Digite a tag: ").strip()
            if tag:
                params['tag'] = tag
        
        if input("Deseja filtrar por source? [s/N]: ").lower() == 's':
            source = input("Digite o ID da fonte: ").strip()
            if source:
                params['source'] = source
    
    if "bookmarks" in endpoint:
        if input("Deseja filtrar por unreadOnly? [s/N]: ").lower() == 's':
            unread = input("Digite true/false: ").strip().lower()
            if unread in ['true', 'false']:
                params['unreadOnly'] = unread == 'true'
        
        if input("Deseja filtrar por listId? [s/N]: ").lower() == 's':
            list_id = input("Digite o ID da lista: ").strip()
            if list_id:
                params['listId'] = list_id
    
    if "posts/{id}/comments" in endpoint:
        if input("Deseja definir ordem (oldest/newest)? [s/N]: ").lower() == 's':
            sort = input("Digite a ordem (oldest/newest): ").strip()
            if sort in ['oldest', 'newest']:
                params['sort'] = sort
    
    if "experiences" in endpoint and "GET" in endpoint:
        if input("Deseja filtrar por tipo? [s/N]: ").lower() == 's':
            print("Opções: work, education, project, certification, volunteering, opensource")
            exp_type = input("Digite o tipo: ").strip()
            if exp_type:
                params['type'] = exp_type
    
    return params


def get_request_body(endpoint, method):
    """
    Solicita dados para o corpo da requisição
    
    Args:
        endpoint: String do endpoint
        method: Método HTTP
    
    Returns:
        Dicionário com os dados do corpo
    """
    if method.upper() not in ["POST", "PUT", "PATCH"]:
        return None
    
    data = {}
    
    if "bookmarks/" in endpoint and method == "POST":
        post_ids = input("Digite os IDs dos posts separados por vírgula: ").strip()
        if post_ids:
            data['postIds'] = [pid.strip() for pid in post_ids.split(',')]
        
        if input("Deseja definir listId? [s/N]: ").lower() == 's':
            list_id = input("Digite o ID da lista: ").strip()
            if list_id:
                data['listId'] = list_id
    
    if "bookmarks/lists" in endpoint and method == "POST":
        name = input("Digite o nome da lista: ").strip()
        if name:
            data['name'] = name
        
        icon = input("Digite o ícone (emoji, opcional): ").strip()
        if icon:
            data['icon'] = icon
    
    if "feeds/custom/" in endpoint and method == "POST":
        name = input("Digite o nome do feed: ").strip()
        if name:
            data['name'] = name
        
        icon = input("Digite o ícone (emoji, opcional): ").strip()
        if icon:
            data['icon'] = icon
        
        if input("Deseja definir orderBy? [s/N]: ").lower() == 's':
            print("Opções: DATE, UPVOTES, DOWNVOTES, COMMENTS, CLICKS")
            order = input("Digite a ordem: ").strip()
            if order:
                data['orderBy'] = order
        
        if input("Deseja definir filtros (minDayRange, minUpvotes, minViews)? [s/N]: ").lower() == 's':
            day_range = input("Digite minDayRange (opcional): ").strip()
            if day_range:
                data['minDayRange'] = int(day_range)
            
            upvotes = input("Digite minUpvotes (opcional): ").strip()
            if upvotes:
                data['minUpvotes'] = int(upvotes)
            
            views = input("Digite minViews (opcional): ").strip()
            if views:
                data['minViews'] = int(views)
    
    if "feeds/filters/tags" in endpoint or "feeds/filters/sources" in endpoint:
        items = input("Digite os nomes/IDs separados por vírgula: ").strip()
        if items:
            key = "tags" if "tags" in endpoint else "sources"
            data[key] = [item.strip() for item in items.split(',')]
    
    if "profile/" in endpoint and method == "PATCH":
        name = input("Digite o nome completo (opcional): ").strip()
        if name:
            data['name'] = name
        
        bio = input("Digite a bio (opcional): ").strip()
        if bio:
            data['bio'] = name
        
        timezone = input("Digite o timezone (opcional): ").strip()
        if timezone:
            data['timezone'] = timezone
        
        if input("Deseja adicionar links sociais? [s/N]: ").lower() == 's':
            social_links = []
            while True:
                url = input("Digite a URL do perfil social (ou deixe vazio para parar): ").strip()
                if not url:
                    break
                platform = input("Digite a plataforma (opcional, será detectada automaticamente): ").strip()
                link = {"url": url}
                if platform:
                    link["platform"] = platform
                social_links.append(link)
            if social_links:
                data['socialLinks'] = social_links
    
    if "profile/stack/" in endpoint and method == "POST":
        title = input("Digite o título da ferramenta: ").strip()
        if title:
            data['title'] = title
        
        print("Seções: primary, hobby, learning, past, custom")
        section = input("Digite a seção: ").strip()
        if section:
            data['section'] = section
        
        if input("Deseja definir startedAt? [s/N]: ").lower() == 's':
            started = input("Digite a data (YYYY-MM-DD): ").strip()
            if started:
                data['startedAt'] = f"{started}T00:00:00Z"
    
    if "profile/experiences/" in endpoint and method in ["POST", "PUT"]:
        print("Tipos: work, education, project, certification, volunteering, opensource")
        exp_type = input("Digite o tipo: ").strip()
        if exp_type:
            data['type'] = exp_type
        
        title = input("Digite o título: ").strip()
        if title:
            data['title'] = title
        
        subtitle = input("Digite o subtítulo (opcional): ").strip()
        if subtitle:
            data['subtitle'] = subtitle
        
        description = input("Digite a descrição (opcional): ").strip()
        if description:
            data['description'] = description
        
        started = input("Digite a data de início (YYYY-MM-DD): ").strip()
        if started:
            data['startedAt'] = f"{started}T00:00:00Z"
        
        if input("Deseja definir data de término? [s/N]: ").lower() == 's':
            ended = input("Digite a data de término (YYYY-MM-DD, ou deixe vazio se atual): ").strip()
            if ended:
                data['endedAt'] = f"{ended}T00:00:00Z"
    
    if "signup" in endpoint:
        email = input("Digite o email: ").strip()
        if email:
            data['email'] = email
        
        password = input("Digite a senha (mínimo 8 caracteres): ").strip()
        if password:
            data['password'] = password
        
        name = input("Digite o nome de exibição (opcional): ").strip()
        if name:
            data['name'] = name
        
        username = input("Digite o handle (opcional): ").strip()
        if username:
            data['username'] = username
    
    return data if data else None


def main():
    # Solicita o bearer token
    token = get_bearer_token()
    
    if not token:
        print("Token não fornecido. Encerrando.")
        return
    
    while True:
        # Exibe o menu
        categories = display_menu()
        
        # Solicita a escolha da categoria
        choice = input("\nDigite o número da categoria (ou 0 para sair): ").strip()
        
        if choice == "0":
            print("Encerrando...")
            break
        
        if choice not in categories:
            print("Opção inválida. Tente novamente.")
            continue
        
        # Exibe as opções da categoria
        cat_name, options = categories[choice]
        print(f"\n--- {cat_name} ---")
        for opt_num, opt_desc in options:
            print(f"{choice}{opt_num}. {opt_desc}")
        
        # Solicita a escolha da opção
        sub_choice = input(f"\nDigite o número da opção (ou 0 para voltar): ").strip()
        
        if sub_choice == "0":
            continue
        
        full_choice = f"{choice}{sub_choice}"
        selected_option = None
        for opt_num, opt_desc in options:
            if opt_num == sub_choice:
                selected_option = opt_desc
                break
        
        if not selected_option:
            print("Opção inválida. Tente novamente.")
            continue
        
        # Extrai método e endpoint da descrição
        parts = selected_option.split(" - ")
        method_endpoint = parts[0]
        method, endpoint = method_endpoint.split(" ", 1)
        
        print(f"\nExecutando: {method} {endpoint}")
        print("-" * 50)
        
        # Trata parâmetros de path
        endpoint, path_params = get_endpoint_params(endpoint)
        
        # Solicita parâmetros de query
        query_params = get_query_params(endpoint)
        
        # Solicita corpo da requisição se necessário
        body_data = get_request_body(endpoint, method)
        
        # Faz a requisição
        result = make_request(method, endpoint, token, query_params, body_data)
        
        if result:
            # Exibe os resultados formatados
            print("\nResultado:")
            print(json.dumps(result, indent=2, ensure_ascii=False))
        else:
            print("Erro na requisição.")
        
        # Pergunta se deseja continuar
        if input("\nDeseja fazer outra requisição? [s/N]: ").lower() != 's':
            print("Encerrando...")
            break


if __name__ == "__main__":
    main()
