import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Account, Transfer } from '../models';
import { MatDialog } from '@angular/material/dialog';
import { ConfirmDialogComponent } from '../core/confirm-dialog.component';
import { v4 as uuidv4 } from 'uuid';

@Component({
  selector: 'app-transfer',
  template: `
    <div>
      <h2 class="bank-title">转账汇款</h2>

      <div class="grid-2">
        <div class="bank-card">
          <h3 class="bank-subtitle">转账信息</h3>

          <div *ngIf="error" class="alert alert-error">{{ error }}</div>
          <div *ngIf="success" class="alert alert-success">转账提交成功！</div>

          <form (ngSubmit)="onSubmit()">
            <div class="bank-form-group">
              <label class="bank-label">转出账户</label>
              <select class="bank-input" [(ngModel)]="form.from_account_id" name="from_account_id" required>
                <option value="">请选择转出账户</option>
                <option *ngFor="let acc of accounts" [value]="acc.id">
                  {{ acc.account_name }} - ¥{{ acc.available_balance | number:'1.2-2' }}
                </option>
              </select>
            </div>

            <div class="bank-form-group">
              <label class="bank-label">转账类型</label>
              <div class="flex gap-2">
                <label style="flex: 1; cursor: pointer;">
                  <input type="radio" [(ngModel)]="form.transfer_type" name="transfer_type" value="intra_bank" (change)="onTypeChange()" />
                  同行转账
                </label>
                <label style="flex: 1; cursor: pointer;">
                  <input type="radio" [(ngModel)]="form.transfer_type" name="transfer_type" value="inter_bank" (change)="onTypeChange()" />
                  跨行转账
                </label>
              </div>
            </div>

            <div class="bank-form-group" *ngIf="form.transfer_type === 'intra_bank'">
              <label class="bank-label">收款账户ID</label>
              <input type="number" class="bank-input" [(ngModel)]="form.to_account_id" name="to_account_id" placeholder="请输入收款账户ID" />
            </div>

            <div class="bank-form-group" *ngIf="form.transfer_type === 'inter_bank'">
              <label class="bank-label">收款银行</label>
              <input type="text" class="bank-input" [(ngModel)]="form.to_bank_name" name="to_bank_name" placeholder="请输入收款银行名称" />
            </div>

            <div class="bank-form-group">
              <label class="bank-label">收款人姓名</label>
              <input type="text" class="bank-input" [(ngModel)]="form.to_account_name" name="to_account_name" required placeholder="请输入收款人姓名" />
            </div>

            <div class="bank-form-group" *ngIf="form.transfer_type === 'inter_bank'">
              <label class="bank-label">收款账号</label>
              <input type="text" class="bank-input" [(ngModel)]="form.to_account_no" name="to_account_no" placeholder="请输入收款账号" />
            </div>

            <div class="bank-form-group">
              <label class="bank-label">转账金额</label>
              <input type="number" class="bank-input" [(ngModel)]="form.amount" name="amount" required min="0.01" step="0.01" placeholder="请输入转账金额" />
              <small *ngIf="calculatedFee >= 0" style="color: #666;">手续费: ¥{{ calculatedFee | number:'1.2-2' }}</small>
            </div>

            <div class="bank-form-group">
              <label class="bank-label">到账方式</label>
              <div class="flex gap-2">
                <label style="flex: 1; cursor: pointer;">
                  <input type="radio" [(ngModel)]="form.transfer_speed" name="transfer_speed" value="realtime" />
                  实时到账
                </label>
                <label style="flex: 1; cursor: pointer;">
                  <input type="radio" [(ngModel)]="form.transfer_speed" name="transfer_speed" value="normal" />
                  普通到账
                </label>
              </div>
            </div>

            <div class="bank-form-group">
              <label class="bank-label">转账附言</label>
              <input type="text" class="bank-input" [(ngModel)]="form.description" name="description" placeholder="请输入附言（选填）" />
            </div>

            <button type="submit" class="bank-btn bank-btn-gold" style="width: 100%;" [disabled]="loading">
              {{ loading ? '处理中...' : '确认转账' }}
            </button>
          </form>
        </div>

        <div>
          <div class="bank-card">
            <h3 class="bank-subtitle">转账限额</h3>
            <div *ngIf="limits" style="font-size: 14px;">
              <div class="flex-between mb-1">
                <span>单笔限额</span>
                <span class="text-primary">¥{{ limits.per_transaction.limit | number:'1.2-2' }}</span>
              </div>
              <div class="flex-between mb-1">
                <span>今日已用</span>
                <span>¥{{ limits.daily.used | number:'1.2-2' }} / ¥{{ limits.daily.limit | number:'1.2-2' }}</span>
              </div>
              <div class="mb-1">
                <div style="height: 8px; background: #eee; border-radius: 4px; overflow: hidden;">
                  <div style="height: 100%; background: var(--primary); width: {{ dailyPercent }}%;"></div>
                </div>
              </div>
              <div class="flex-between mb-1">
                <span>本月已用</span>
                <span>¥{{ limits.monthly.used | number:'1.2-2' }} / ¥{{ limits.monthly.limit | number:'1.2-2' }}</span>
              </div>
              <div>
                <div style="height: 8px; background: #eee; border-radius: 4px; overflow: hidden;">
                  <div style="height: 100%; background: var(--accent-gold); width: {{ monthlyPercent }}%;"></div>
                </div>
              </div>
            </div>
          </div>

          <div class="bank-card mt-2">
            <h3 class="bank-subtitle">温馨提示</h3>
            <ul style="color: #666; font-size: 13px; padding-left: 20px;">
              <li class="mb-1">同行转账实时到账，跨行转账需2-3秒处理</li>
              <li class="mb-1">单次转账超过5万元可能触发风控审核</li>
              <li class="mb-1">转账前请仔细核对收款人信息</li>
              <li class="mb-1">银行不会向您索要验证码和密码</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  `
})
export class TransferComponent implements OnInit {
  accounts: Account[] = [];
  limits: any = null;
  dailyPercent = 0;
  monthlyPercent = 0;

  form = {
    from_account_id: null as any,
    to_account_id: null as any,
    to_account_no: '',
    to_bank_name: '',
    to_account_name: '',
    amount: 0,
    transfer_type: 'intra_bank',
    transfer_speed: 'realtime',
    description: '',
    biz_id: ''
  };

  calculatedFee = 0;
  error = '';
  success = false;
  loading = false;

  private apiUrl = 'http://localhost:8080/api';

  constructor(
    private http: HttpClient,
    private dialog: MatDialog
  ) {}

  ngOnInit(): void {
    this.loadAccounts();
    this.loadLimits();
  }

  loadAccounts(): void {
    this.http.get<Account[]>(`${this.apiUrl}/accounts`).subscribe({
      next: (data) => {
        this.accounts = data.filter(a => a.status === 'active');
      }
    });
  }

  loadLimits(): void {
    this.http.get(`${this.apiUrl}/limits?scope=transfer`).subscribe({
      next: (data) => {
        this.limits = data;
        this.dailyPercent = (this.limits.daily.used / this.limits.daily.limit) * 100;
        this.monthlyPercent = (this.limits.monthly.used / this.limits.monthly.limit) * 100;
      }
    });
  }

  onTypeChange(): void {
    if (this.form.transfer_type === 'intra_bank') {
      this.form.to_account_no = '';
      this.form.to_bank_name = '';
    } else {
      this.form.to_account_id = null;
    }
    this.calculateFee();
  }

  calculateFee(): void {
    if (!this.form.amount) {
      this.calculatedFee = 0;
      return;
    }
    let rate = 0.001;
    if (this.form.transfer_type === 'inter_bank') {
      rate = 0.002;
    }
    if (this.form.transfer_speed === 'realtime') {
      rate += 0.001;
    }
    let fee = this.form.amount * rate;
    const maxFee = 50;
    const minFee = this.form.transfer_type === 'inter_bank' ? 2 : 0;
    fee = Math.max(minFee, Math.min(maxFee, fee));
    this.calculatedFee = Math.round(fee * 100) / 100;
  }

  onSubmit(): void {
    this.error = '';
    this.success = false;

    if (!this.form.from_account_id) {
      this.error = '请选择转出账户';
      return;
    }
    if (!this.form.amount || this.form.amount <= 0) {
      this.error = '请输入正确的转账金额';
      return;
    }
    if (!this.form.to_account_name) {
      this.error = '请输入收款人姓名';
      return;
    }
    if (this.form.transfer_type === 'intra_bank' && !this.form.to_account_id) {
      this.error = '请输入收款账户ID';
      return;
    }
    if (this.form.transfer_type === 'inter_bank' && !this.form.to_account_no) {
      this.error = '请输入收款账号';
      return;
    }

    const dialogRef = this.dialog.open(ConfirmDialogComponent, {
      data: {
        title: '确认转账',
        message: `确定向 ${this.form.to_account_name} 转账 ¥${this.form.amount.toFixed(2)} 吗？`
      }
    });

    dialogRef.afterClosed().subscribe((result: boolean) => {
      if (result) {
        this.doTransfer();
      }
    });
  }

  doTransfer(): void {
    this.loading = true;
    this.form.biz_id = uuidv4();

    this.http.post<Transfer>(`${this.apiUrl}/transfers`, this.form).subscribe({
      next: () => {
        this.loading = false;
        this.success = true;
        this.loadAccounts();
        this.loadLimits();

        setTimeout(() => {
          this.success = false;
        }, 3000);
      },
      error: (err) => {
        this.loading = false;
        this.error = err.error?.error || '转账失败，请重试';
      }
    });
  }
}
