// Gera a colecao do Insomnia (export format v4) com todas as rotas da plataforma.
const fs = require("fs");

const out = process.argv[2];
const now = Date.now();
const WS = "wrk_food_delivery";

// Tag de resposta do Insomnia: le um campo (JSONPath) da ultima resposta de outra request.
const b64 = (s) => Buffer.from(s).toString("base64");
// Script after-response: grava campos do body (2xx) no environment ativo.
// map = { variavel_do_env: "campo_do_body" }
const saveScript = (map) =>
  [
    "// Salva campos da resposta no environment ativo.",
    "const status = insomnia.response.code;",
    "if (status >= 200 && status < 300) {",
    "  const body = insomnia.response.json();",
    ...Object.entries(map).map(
      ([envVar, field]) => `  insomnia.environment.set("${envVar}", body.${field});`
    ),
    "}",
    "",
  ].join("\n");

const resources = [];
let sort = 0;

resources.push({
  _id: WS,
  _type: "workspace",
  parentId: null,
  name: "Food Delivery Platform",
  description: "Rotas do auth-service, restaurant-service e api-gateway.",
  scope: "collection",
  created: now,
  modified: now,
});

resources.push({
  _id: "env_base",
  _type: "environment",
  parentId: WS,
  name: "Base Environment",
  data: {
    auth_url: "http://localhost:8081",
    restaurant_url: "http://localhost:8083",
    gateway_url: "http://localhost:8080",
    email: "dev@fooddelivery.local",
    password: "Senha@Forte123",
    // Preenchidos pelos scripts after-response de Login, Criar restaurante e Criar item.
    // (Tags {% response %} nao servem: o Insomnia gera ids novos ao importar.)
    access_token: "",
    refresh_token: "",
    restaurant_id: "",
    item_id: "",
  },
  dataPropertyOrder: {
    "&": [
      "auth_url", "restaurant_url", "gateway_url", "email", "password",
      "access_token", "refresh_token", "restaurant_id", "item_id",
    ],
  },
  color: null,
  isPrivate: false,
  metaSortKey: 0,
  created: now,
  modified: now,
});

// Sub-environment com os mesmos valores: na importacao o Insomnia pode manter o
// Base Environment ja existente da colecao (vazio) e descartar o do arquivo;
// sub-environments sempre sao criados.
{
  const base = resources.find((r) => r._id === "env_base");
  resources.push({
    ...base,
    _id: "env_local",
    parentId: "env_base",
    name: "Local",
    data: { ...base.data },
    color: "#7d69cb",
    metaSortKey: 1,
  });
}

resources.push({
  _id: "jar_food_delivery",
  _type: "cookie_jar",
  parentId: WS,
  name: "Default Jar",
  cookies: [],
  created: now,
  modified: now,
});

function folder(id, parentId, name, description = "") {
  resources.push({
    _id: id,
    _type: "request_group",
    parentId,
    name,
    description,
    environment: {},
    environmentPropertyOrder: null,
    metaSortKey: ++sort,
    created: now,
    modified: now,
  });
}

function request(id, parentId, name, method, url, opts = {}) {
  const headers = [];
  let body = {};
  if (opts.json !== undefined) {
    headers.push({ name: "Content-Type", value: "application/json" });
    body = { mimeType: "application/json", text: JSON.stringify(opts.json, null, 2) };
  }
  for (const [name, value] of Object.entries(opts.headers || {})) headers.push({ name, value });

  resources.push({
    _id: id,
    _type: "request",
    parentId,
    name,
    description: opts.description || "",
    method,
    url,
    body,
    parameters: (opts.query || []).map(([name, value]) => ({ name, value, disabled: false })),
    headers,
    authentication: opts.bearer
      ? { type: "bearer", token: "{{ _.access_token }}", prefix: "" }
      : {},
    afterResponseScript: opts.saveToEnv ? saveScript(opts.saveToEnv) : "",
    settingStoreCookies: true,
    settingSendCookies: true,
    settingDisableRenderRequestBody: false,
    settingEncodeUrl: true,
    settingRebuildPath: true,
    settingFollowRedirects: "global",
    metaSortKey: ++sort,
    isPrivate: false,
    created: now,
    modified: now,
  });
}

const A = "{{ _.auth_url }}";
const R = "{{ _.restaurant_url }}";
const G = "{{ _.gateway_url }}";
const newKey = "{% uuid 'v4' %}";

// ---------------- Auth Service ----------------
folder("fld_auth", WS, "Auth Service", "Porta 8081. Emite os JWT aceitos pelas rotas admin do restaurant-service.");

folder("fld_auth_health", "fld_auth", "Health");
request("req_auth_live", "fld_auth_health", "Live", "GET", `${A}/health/live`);
request("req_auth_ready", "fld_auth_health", "Ready", "GET", `${A}/health/ready`);
request("req_auth_health", "fld_auth_health", "Auth health", "GET", `${A}/auth/health`);
request("req_auth_metrics", "fld_auth_health", "Metrics (Prometheus)", "GET", `${A}/metrics`);

folder("fld_auth_flow", "fld_auth", "Autenticacao");
const SAVE_TOKENS = { access_token: "access_token", refresh_token: "refresh_token" };

request("req_auth_register", "fld_auth_flow", "1. Register", "POST", `${A}/auth/register`, {
  description: "Cria o usuario do environment e ja salva os tokens. Se ja existir (409), siga para o Login.",
  saveToEnv: SAVE_TOKENS,
  json: { email: "{{ _.email }}", password: "{{ _.password }}" },
  headers: { "Idempotency-Key": newKey },
});
request("req_auth_login", "fld_auth_flow", "2. Login", "POST", `${A}/auth/login`, {
  description: "Salva access_token e refresh_token no environment ativo (aba Scripts > After-response). Rode de novo quando o token expirar (15 min).",
  saveToEnv: SAVE_TOKENS,
  json: { email: "{{ _.email }}", password: "{{ _.password }}" },
  headers: { "Idempotency-Key": newKey },
});
request("req_auth_refresh", "fld_auth_flow", "3. Refresh", "POST", `${A}/auth/refresh`, {
  description: "Troca o refresh_token por um par novo e atualiza as duas variaveis. Depois de um Logout, faca Login de novo.",
  saveToEnv: SAVE_TOKENS,
  json: { refresh_token: "{{ _.refresh_token }}" },
});
request("req_auth_logout", "fld_auth_flow", "4. Logout", "POST", `${A}/auth/logout`, {
  description: "Revoga o refresh_token do ultimo Login (204).",
  json: { refresh_token: "{{ _.refresh_token }}" },
});

// ---------------- Restaurant Service ----------------
folder("fld_rest", WS, "Restaurant Service", "Porta 8083. Rotas admin exigem o access_token do auth-service.");

folder("fld_rest_health", "fld_rest", "Health e Metrics");
request("req_rest_live", "fld_rest_health", "Live", "GET", `${R}/health/live`);
request("req_rest_ready", "fld_rest_health", "Ready", "GET", `${R}/health/ready`);
request("req_rest_metrics", "fld_rest_health", "Metrics (Prometheus)", "GET", `${R}/metrics`);

folder("fld_rest_admin", "fld_rest", "Admin (JWT)", "Rode Auth Service > Login antes.");
request("req_rest_create", "fld_rest_admin", "1. Criar restaurante", "POST", `${R}/restaurants`, {
  description: "Salva o id criado em restaurant_id do environment ativo.",
  saveToEnv: { restaurant_id: "id" },
  bearer: true,
  json: {
    name: "Pizzaria Bella Napoli",
    address: "Rua das Flores, 123 - Sao Paulo",
    phone_number: "11999999999",
    delivery_fee: 7.5,
  },
  headers: { "Idempotency-Key": newKey },
});
request("req_item_create", "fld_rest_admin", "2. Criar item do cardapio", "POST", `${R}/restaurants/{{ _.restaurant_id }}/menu/items`, {
  description:
    "Salva o id criado em item_id do environment ativo. Categorias validas: pizzaria, hamburgueria, japones, comida brasileira, sorveteria.",
  saveToEnv: { item_id: "id" },
  bearer: true,
  json: { name: "Pizza Margherita", price: 49.9, category: "pizzaria", available: true },
  headers: { "Idempotency-Key": newKey },
});
request("req_item_update", "fld_rest_admin", "3. Atualizar item (PATCH)", "PATCH", `${R}/restaurants/{{ _.restaurant_id }}/menu/items/{{ _.item_id }}`, {
  description: "Todos os campos sao opcionais; envie so o que mudar.",
  bearer: true,
  json: { price: 54.9, available: false },
});

folder("fld_rest_public", "fld_rest", "Publico", "Sem autenticacao, com rate limit por IP.");
request("req_rest_list", "fld_rest_public", "Listar restaurantes", "GET", `${R}/restaurants`, {
  description: "page ate 1000; page_size entre 1 e 100 (default 20).",
  query: [["page", "1"], ["page_size", "20"]],
});
request("req_rest_get", "fld_rest_public", "Detalhes do restaurante", "GET", `${R}/restaurants/{{ _.restaurant_id }}`);
request("req_rest_menu", "fld_rest_public", "Cardapio do restaurante", "GET", `${R}/restaurants/{{ _.restaurant_id }}/menu`);

folder("fld_rest_internal", "fld_rest", "Interno (order-service)", "Sem autenticacao por enquanto (servico-a-servico pendente).");
request("req_rest_validate", "fld_rest_internal", "Validar itens do pedido", "POST", `${R}/restaurants/{{ _.restaurant_id }}/menu/validate`, {
  description: "Ate 100 item_ids. O segundo id nao existe, para mostrar item_not_found.",
  json: { item_ids: ["{{ _.item_id }}", "000000000000000000000000"] },
});

folder("fld_rest_errors", "fld_rest", "Cenarios de erro", "Respostas esperadas no nome de cada request.");
request("req_err_no_token", "fld_rest_errors", "401 - Admin sem token", "POST", `${R}/restaurants`, {
  json: { name: "Sem token", address: "Rua X, 1", phone_number: "11900000000", delivery_fee: 1 },
});
request("req_err_page_size", "fld_rest_errors", "400 - page_size acima de 100", "GET", `${R}/restaurants`, {
  query: [["page", "1"], ["page_size", "500"]],
});
request("req_err_category", "fld_rest_errors", "400 - Categoria invalida", "POST", `${R}/restaurants/{{ _.restaurant_id }}/menu/items`, {
  bearer: true,
  json: { name: "Taco", price: 20, category: "mexicana" },
  headers: { "Idempotency-Key": newKey },
});
request("req_err_idem_retry", "fld_rest_errors", "201 - Retry com a mesma Idempotency-Key", "POST", `${R}/restaurants/{{ _.restaurant_id }}/menu/items`, {
  description: "Envie duas vezes: a segunda devolve o mesmo item. Troque a chave para criar outro.",
  bearer: true,
  json: { name: "Pizza Calabresa", price: 52.9, category: "pizzaria" },
  headers: { "Idempotency-Key": "insomnia-retry-demo-1" },
});
request("req_err_not_found", "fld_rest_errors", "404 - Restaurante inexistente", "GET", `${R}/restaurants/00000000-0000-0000-0000-000000000000`);

// ---------------- API Gateway ----------------
folder("fld_gw", WS, "API Gateway", "Porta 8080. Ainda nao esta no Docker Compose: suba com go run ./cmd/api em api-gateway/. So faz proxy das rotas de auth.");
request("req_gw_live", "fld_gw", "Live", "GET", `${G}/health/live`);
request("req_gw_ready", "fld_gw", "Ready", "GET", `${G}/health/ready`);
request("req_gw_register", "fld_gw", "Register (proxy)", "POST", `${G}/auth/register`, {
  saveToEnv: SAVE_TOKENS,
  json: { email: "{{ _.email }}", password: "{{ _.password }}" },
});
request("req_gw_login", "fld_gw", "Login (proxy)", "POST", `${G}/auth/login`, {
  saveToEnv: SAVE_TOKENS,
  json: { email: "{{ _.email }}", password: "{{ _.password }}" },
});
request("req_gw_refresh", "fld_gw", "Refresh (proxy)", "POST", `${G}/auth/refresh`, {
  saveToEnv: SAVE_TOKENS,
  json: { refresh_token: "{{ _.refresh_token }}" },
});
request("req_gw_logout", "fld_gw", "Logout (proxy)", "POST", `${G}/auth/logout`, {
  json: { refresh_token: "{{ _.refresh_token }}" },
});

const doc = {
  _type: "export",
  __export_format: 4,
  __export_date: new Date(now).toISOString(),
  __export_source: "insomnia.desktop.app:v10.0.0",
  resources,
};
fs.writeFileSync(out, JSON.stringify(doc, null, 2) + "\n");
console.log(`${resources.filter((r) => r._type === "request").length} requests -> ${out}`);
