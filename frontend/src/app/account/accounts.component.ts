import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Account } from '../models';
import { MatDialog } from '@angular/material/dialog';
import { ConfirmDialogComponent } from '../core/confirm-dialog.component';
import { AuthService } from '../core/auth.service';

@Component({
  selector: 'app-accounts',
  template: `
    <div>
      <div class="flex-between mb-3">
        <h2 class="bank-title" style="margin: 0;">我的账户</h2>
        <button class="bank-btn bank-btn-gold" (click)="openCreateDialog()">
          + 开立账户
        </button>
      </div>

      <div *ngIf="accounts.length === 0" class="bank-card" style="text-align: center; padding: 60px;">
        <p style="color: #999; font-size: 16px;">还没有账户，点击右上角开立您的第一个账户</p>
      </div>

      <div class="grid-3">
        <div *ngFor="let account of accounts" class="bank-card" style="background: linear-gradient(135deg, var(--primary-dark), var(--primary-light)); color: white;">
          <div class="flex-between mb-2">
            <span style="font-size: 14px; opacity: 0.8;">{{ getAccountTypeName(account.account_type) }}</span>
            <span class="bank-badge" [ngClass]="getStatusClass(account.status)">{{ getStatusName(account.status) }}</span>
          </div>
          <div style="font-size: 14px; opacity: 0.8; margin-bottom: 8px;">
            {{ account.account_number }}
          </div>
          <div style="font-size: 28px; font-weight: 700; margin-bottom: 8px;">
            ¥{{ account.balance | number:'1.2-2' }}
          </div>
          <div style="font-size: 13px; opacity: 0.7; margin-bottom: 16px;">
            可用余额: ¥{{ account.available_balance | number:'1.2-2' }}
            <span *ngIf="account.frozen_amount > 0" style="margin-left: 12px;">
              冻结: ¥{{ account.frozen_amount | number:'1.2-2' }}
            </span>
          </div>
          <div style="font-size: 12px; opacity: 0.6; margin-bottom: 16px;">
            年利率: {{ (account.interest_rate * 100).toFixed(2) }}%
            <span *ngIf="account.maturity_date" style="margin-left: 12px;">
              到期日: {{ account.maturity_date | date:'yyyy-MM-dd' }}
            </span>
          </div>
          <div class="flex gap-1">
            <button class="bank-btn" style="flex: 1; padding: 8px; font-size: 13px; background: rgba(255,255,255,0.2); color: white;" 
                    (click)="viewDetail(account)">详情</button>
            <button class="bank-btn" style="flex: 1; padding: 8px; font-size: 13px; background: var(--accent-gold); color: #1a1a1a;"
                    *ngIf="account.status === 'active'" (click)="freezeAccount(account)">冻结</button>
            <button class="bank-btn" style="flex: 1; padding: 8px; font-size: 13px; background: var(--accent-gold); color: #1a1a1a;"
                    *ngIf="account.status === 'frozen'" (click)="unfreezeAccount(account)">解冻</button>
          </div>
        </div>
      </div>

      <div *ngIf="selectedAccount" class="bank-card mt-3">
        <h3 class="bank-subtitle">账户明细 - {{ selectedAccount.account_name }}</h3>
        <table class="bank-table">
          <thead>
            <tr>
              <th>时间</th>
              <th>类型</th>
              <th>描述</th>
              <th style="text-align: right;">金额</th>
              <th style="text-align: right;">余额</th>
            </tr>
          </thead>
          <tbody>
            <tr *ngFor="let entry of ledgerEntries">
              <td>{{ entry.created_at | date:'yyyy-MM-dd HH:mm' }}</td>
              <td>
                <span class="bank-badge" [ngClass]="entry.entry_type === 'credit' ? 'bank-badge-success' : 'bank-badge-danger'">
                  {{ entry.entry_type === 'credit' ? '贷' : '借' }}
                </span>
              </td>
              <td>{{ entry.description }}</td>
              <td style="text-align: right;" [ngClass]="entry.entry_type === 'credit' ? 'amount-positive' : 'amount-negative'">
                {{ entry.entry_type === 'credit' ? '+' : '-' }}¥{{ entry.amount | number:'1.2-2' }}
              </td>
              <td style="text-align: right;">¥{{ entry.balance_after | number:'1.2-2' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div *ngIf="showCreateDialog" class="bank-modal-overlay" (click.self)="showCreateDialog = false">
        <div class="bank-modal">
          <div class="bank-modal-header">
            <span>开立新账户</span>
            <button class="bank-modal-close" (click)="showCreateDialog = false">&times;</button>
          </div>
          <form (ngSubmit)="createAccount()">
            <div class="bank-form-group">
              <label class="bank-label">账户类型</label>
              <select class="bank-input" [(ngModel)]="newAccount.account_type" name="account_type" required>
                <option value="savings">活期储蓄账户</option>
                <option value="checking">支票账户</option>
                <option value="fixed_deposit">定期存款账户</option>
              </select>
            </div>
            <div class="bank-form-group">
              <label class="bank-label">账户名称</label>
              <input type="text" class="bank-input" [(ngModel)]="newAccount.account_name" name="account_name" required placeholder="如：工资卡、零花钱" />
            </div>
            <div class="bank-form-group">
              <label class="bank-label">币种</label>
              <select class="bank-input" [(ngModel)]="newAccount.currency" name="currency">
                <option value="CNY">人民币 (CNY)</option>
                <option value="USD">美元 (USD)</option>
                <option value="EUR">欧元 (EUR)</option>
              </select>
            </div>
            <div class="bank-form-group" *ngIf="newAccount.account_type === 'fixed_deposit'">
              <label class="bank-label">存期（天）</label>
              <input type="number" class="bank-input" [(ngModel)]="newAccount.term_days" name="term_days" min="7" placeholder="如：90、180、365" />
            </div>
            <div class="bank-form-group">
              <label class="bank-label">初始存入金额</label>
              <input type="number" class="bank-input" [(ngModel)]="newAccount.amount" name="amount" min="0" step="0.01" />
            </div>
            <div class="bank-modal-footer">
              <button type="button" class="bank-btn bank-btn-outline" (click)="showCreateDialog = false">取消</button>
              <button type="submit" class="bank-btn bank-btn-gold">确认开立</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  `
})
export class AccountsComponent implements OnInit {
  accounts: Account[] = [];
  selectedAccount: Account | null = null;
  ledgerEntries: any[] = [];
  showCreateDialog = false;
  newAccount = {
    account_type: 'savings',
    account_name: '',
    currency: 'CNY',
    amount: 0,
    term_days: 0
  };

  private apiUrl = 'http://localhost:8080/api';

  constructor(
    private http: HttpClient,
    private dialog: MatDialog,
    private authService: AuthService
  ) {}

  ngOnInit(): void {
    this.loadAccounts();
  }

  loadAccounts(): void {
    this.http.get<Account[]>(`${this.apiUrl}/accounts`).subscribe({
      next: (data) => {
        this.accounts = data;
      },
      error: (err) => {
        console.error('Failed to load accounts', err);
      }
    });
  }

  viewDetail(account: Account): void {
    this.selectedAccount = account;
    this.http.get(`${this.apiUrl}/accounts/${account.id}/ledger?page_size=20`).subscribe({
      next: (data: any) => {
        this.ledgerEntries = data.entries || [];
      }
    });
  }

  openCreateDialog(): void {
    this.showCreateDialog = true;
    this.newAccount = {
      account_type: 'savings',
      account_name: '',
      currency: 'CNY',
      amount: 0,
      term_days: 0
    };
  }

  createAccount(): void {
    if (!this.newAccount.account_name) return;

    this.http.post(`${this.apiUrl}/accounts`, this.newAccount).subscribe({
      next: () => {
        this.showCreateDialog = false;
        this.loadAccounts();
      },
      error: (err) => {
        alert(err.error?.error || '开户失败');
      }
    });
  }

  freezeAccount(account: Account): void {
    const dialogRef = this.dialog.open(ConfirmDialogComponent, {
      data: {
        title: '确认冻结',
        message: `确定要冻结账户 "${account.account_name}" 吗？冻结后账户将无法进行交易。`
      }
    });

    dialogRef.afterClosed().subscribe((result: boolean) => {
      if (result) {
        this.http.post(`${this.apiUrl}/accounts/${account.id}/freeze`, { reason: '用户主动冻结' }).subscribe({
          next: () => {
            this.loadAccounts();
          }
        });
      }
    });
  }

  unfreezeAccount(account: Account): void {
    const dialogRef = this.dialog.open(ConfirmDialogComponent, {
      data: {
        title: '确认解冻',
        message: `确定要解冻账户 "${account.account_name}" 吗？`
      }
    });

    dialogRef.afterClosed().subscribe((result: boolean) => {
      if (result) {
        this.http.post(`${this.apiUrl}/accounts/${account.id}/unfreeze`, {}).subscribe({
          next: () => {
            this.loadAccounts();
          }
        });
      }
    });
  }

  getAccountTypeName(type: string): string {
    const map: Record<string, string> = {
      'savings': '活期储蓄',
      'fixed_deposit': '定期存款',
      'checking': '支票账户'
    };
    return map[type] || type;
  }

  getStatusName(status: string): string {
    const map: Record<string, string> = {
      'active': '正常',
      'frozen': '已冻结',
      'closed': '已销户'
    };
    return map[status] || status;
  }

  getStatusClass(status: string): string {
    const map: Record<string, string> = {
      'active': 'bank-badge-success',
      'frozen': 'bank-badge-warning',
      'closed': 'bank-badge-danger'
    };
    return map[status] || 'bank-badge-info';
  }
}
