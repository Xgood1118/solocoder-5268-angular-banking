import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ReconReport, ReconDifference, Account } from '../models';

@Component({
  selector: 'app-recon',
  template: `
    <div>
      <h2 class="bank-title">对账中心</h2>

      <div class="bank-card">
        <div class="flex-between mb-3">
          <h3 class="bank-subtitle" style="margin: 0;">对账报告</h3>
          <div class="flex gap-2">
            <select class="bank-input" style="width: 200px;" [(ngModel)]="selectedAccount">
              <option value="">全部账户</option>
              <option *ngFor="let acc of accounts" [value]="acc.id">{{ acc.account_name }}</option>
            </select>
            <button class="bank-btn bank-btn-primary" (click)="triggerRecon()">
              立即对账
            </button>
          </div>
        </div>

        <table class="bank-table">
          <thead>
            <tr>
              <th>对账日期</th>
              <th>账户</th>
              <th>系统余额</th>
              <th>流水余额</th>
              <th>差异金额</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr *ngFor="let report of reports">
              <td>{{ report.recon_date }}</td>
              <td>{{ report.account_no }}</td>
              <td>¥{{ report.system_balance | number:'1.2-2' }}</td>
              <td>¥{{ report.ledger_balance | number:'1.2-2' }}</td>
              <td [ngClass]="report.difference > 0 ? 'amount-negative' : 'amount-positive'">
                ¥{{ report.difference | number:'1.2-2' }}
              </td>
              <td>
                <span class="bank-badge" [ngClass]="report.status === 'success' ? 'bank-badge-success' : 'bank-badge-warning'">
                  {{ report.status === 'success' ? '对账成功' : '存在差异' }}
                </span>
              </td>
              <td>
                <button *ngIf="report.status !== 'success'" 
                        class="bank-btn bank-btn-outline" 
                        style="padding: 4px 12px; font-size: 12px;"
                        (click)="viewDifferences(report)">
                  查看差异
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div *ngIf="showDiffModal" class="bank-modal-overlay" (click.self)="showDiffModal = false">
        <div class="bank-modal" style="max-width: 700px;">
          <div class="bank-modal-header">
            <span>对账差异明细</span>
            <button class="bank-modal-close" (click)="showDiffModal = false">&times;</button>
          </div>
          <table class="bank-table">
            <thead>
              <tr>
                <th>差异类型</th>
                <th>交易ID</th>
                <th>预期金额</th>
                <th>实际金额</th>
                <th>说明</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              <tr *ngFor="let diff of differences">
                <td>
                  <span class="bank-badge bank-badge-warning">
                    {{ getDiffTypeName(diff.diff_type) }}
                  </span>
                </td>
                <td>{{ diff.transaction_id }}</td>
                <td>¥{{ diff.expected_amount | number:'1.2-2' }}</td>
                <td>¥{{ diff.actual_amount | number:'1.2-2' }}</td>
                <td>{{ diff.description }}</td>
                <td>
                  <span *ngIf="diff.is_manual_recon" class="bank-badge bank-badge-success">已人工对账</span>
                  <span *ngIf="!diff.is_manual_recon" class="bank-badge bank-badge-danger">待处理</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `
})
export class ReconComponent implements OnInit {
  reports: ReconReport[] = [];
  accounts: Account[] = [];
  differences: ReconDifference[] = [];
  showDiffModal = false;
  selectedAccount = '';

  private apiUrl = 'http://localhost:8080/api';

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.loadAccounts();
    this.loadReports();
  }

  loadAccounts(): void {
    this.http.get<Account[]>(`${this.apiUrl}/accounts`).subscribe({
      next: (data) => {
        this.accounts = data;
      }
    });
  }

  loadReports(): void {
    let url = `${this.apiUrl}/recon/reports`;
    if (this.selectedAccount) {
      url += `?account_id=${this.selectedAccount}`;
    }

    this.http.get<any>(url).subscribe({
      next: (data) => {
        this.reports = data.reports || [];
      }
    });
  }

  triggerRecon(): void {
    if (!this.selectedAccount) {
      alert('请选择账户');
      return;
    }

    this.http.post(`${this.apiUrl}/recon/trigger`, { account_id: Number(this.selectedAccount) }).subscribe({
      next: () => {
        this.loadReports();
      }
    });
  }

  viewDifferences(report: ReconReport): void {
    this.http.get<any>(`${this.apiUrl}/recon/reports/${report.id}/differences`).subscribe({
      next: (data) => {
        this.differences = data.differences || [];
        this.showDiffModal = true;
      }
    });
  }

  getDiffTypeName(type: string): string {
    const map: Record<string, string> = {
      'amount_mismatch': '金额不符',
      'missing_entry': '缺少流水',
      'extra_entry': '多余流水'
    };
    return map[type] || type;
  }
}
