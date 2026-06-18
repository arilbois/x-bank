// Mirrors the Go backend JSON shapes in internal/models and internal/handlers.

export const CATEGORIES = ["sambatWarga", "persibWay", "bytmod"] as const;
export type Category = (typeof CATEGORIES)[number];

export const STATUSES = [
  "scraped",
  "analyzing",
  "analyzed",
  "failed",
] as const;
export type ArticleStatus = (typeof STATUSES)[number];

export interface Article {
  id: string;
  source_category: Category;
  source_name: string;
  title: string;
  url: string;
  excerpt: string;
  author: string;
  image_url: string;
  published_at: string | null;
  scraped_at: string;
  score: number;
  status: ArticleStatus;
  content_hash: string;
  tags: string[] | null;
  created_at: string;
  updated_at: string;
}

export interface ArticleAnalysis {
  id: string;
  article_id: string;
  summary: string;
  sentiment: "positive" | "negative" | "neutral" | string;
  hook: string;
  tweet: string;
  thread_opener: string;
  model: string;
  tokens_used: number;
  created_at: string;
}

export interface User {
  id: string;
  username: string;
  role: "admin" | "user" | string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ListArticlesResponse {
  data: Article[];
  total: number;
  page: number;
  limit: number;
}

export interface SingleArticleResponse {
  data: Article;
}

export interface TrendingResponse {
  data: Article[];
  count: number;
}

export interface ScrapeRunResponse {
  status: string;
}

export interface ApiError {
  error: string;
}
