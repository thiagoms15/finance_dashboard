export type User = {
  id: string;
  name: string;
  email: string;
  createdAt: string;
  updatedAt: string;
};

export type Asset = {
  id: string;
  symbol: string;
  name: string;
  exchange: string;
  currency: string;
  sector: string;
  createdAt: string;
  updatedAt: string;
};

export type Transaction = {
  id: string;
  userId: string;
  assetId: string;
  type: "BUY" | "SELL" | "DIVIDEND";
  quantity: string;
  price: string;
  fees: string;
  currency: string;
  transactionDate: string;
  notes: string;
  createdAt: string;
  updatedAt: string;
};

export type Dividend = {
  id: string;
  userId: string;
  assetId: string;
  amount: string;
  currency: string;
  paymentDate: string;
  createdAt: string;
  updatedAt: string;
};

export type PortfolioPosition = {
  asset: Asset;
  quantity: string;
  averageCost: string;
  totalCost: string;
  currentPrice: string;
  currentValue: string;
  currentCurrency: string;
  unrealizedPL: string;
  unrealizedPLPct: string;
  dailyChange: string;
  dailyChangePct: string;
  realizedPL: string;
};

export type PortfolioSummary = {
  preferredCurrency: string;
  totalInvested: string;
  currentValue: string;
  totalProfitLoss: string;
  dailyGainLoss: string;
  realizedProfit: string;
  dividendsReceived: string;
};

export type PortfolioSnapshot = {
  positions: PortfolioPosition[];
  summary: PortfolioSummary;
};

export type PerformancePoint = {
  date: string;
  value: string;
};

export type LoginResponse = {
  accessToken: string;
  tokenType: string;
  expiresIn: number;
  user: User;
};

export type ApiListResponse<T> = {
  data: T[];
};

export type CreateAssetRequest = {
  symbol: string;
  name: string;
  exchange: string;
  currency: string;
  sector: string;
};
