export interface User {
  id: number;
  username: string;
  email: string;
  phone: string;
  full_name: string;
  id_card: string;
  status: string;
  twofa_enabled: boolean;
  created_at: string;
}

export interface Account {
  id: number;
  user_id: number;
  account_number: string;
  account_type: string;
  account_name: string;
  currency: string;
  status: string;
  balance: number;
  available_balance: number;
  frozen_amount: number;
  interest_rate: number;
  term_days: number;
  maturity_date?: string;
  created_at: string;
}

export interface LedgerEntry {
  id: number;
  account_id: number;
  transaction_id: string;
  biz_id: string;
  entry_type: string;
  amount: number;
  balance_after: number;
  description: string;
  ref_account_id: number;
  created_at: string;
}

export interface Transfer {
  id: number;
  biz_id: string;
  user_id: number;
  from_account_id: number;
  to_account_id: number;
  from_account_no: string;
  to_account_no: string;
  to_bank_name: string;
  to_account_name: string;
  amount: number;
  currency: string;
  transfer_type: string;
  transfer_speed: string;
  status: string;
  description: string;
  fee: number;
  clearing_ref_no: string;
  failure_reason: string;
  created_at: string;
  completed_at?: string;
}

export interface LoginResponse {
  token?: string;
  user?: User;
  need_twofa?: boolean;
  twofa_token?: string;
}

export interface AuditLog {
  id: number;
  user_id: number;
  action: string;
  module: string;
  description: string;
  ip_address: string;
  created_at: string;
}

export interface ReconReport {
  id: number;
  recon_date: string;
  account_id: number;
  account_no: string;
  system_balance: number;
  ledger_balance: number;
  difference: number;
  status: string;
  total_entries: number;
  created_at: string;
}

export interface ReconDifference {
  id: number;
  report_id: number;
  account_id: number;
  diff_type: string;
  transaction_id: string;
  expected_amount: number;
  actual_amount: number;
  description: string;
  is_manual_recon: boolean;
  created_at: string;
}

export interface LimitInfo {
  limit: number;
  used: number;
}
