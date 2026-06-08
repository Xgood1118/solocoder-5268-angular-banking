import { Component, Inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';

@Component({
  selector: 'app-twofa-dialog',
  template: `
    <div class="bank-modal">
      <div class="bank-modal-header">
        <span>双因子认证</span>
        <button class="bank-modal-close" (click)="onCancel()">&times;</button>
      </div>
      <div class="bank-modal-body">
        <p class="mb-2">验证码已发送至 {{ data.target }}</p>
        <p class="mb-2" style="color: #666; font-size: 13px;">
          请输入6位数字验证码，5分钟内有效
        </p>
        <div class="bank-form-group">
          <label class="bank-label">验证码</label>
          <input
            type="text"
            class="bank-input"
            [(ngModel)]="code"
            maxlength="6"
            placeholder="请输入6位验证码"
            style="font-size: 24px; letter-spacing: 8px; text-align: center;"
          />
        </div>
        <div *ngIf="error" class="alert alert-error">
          {{ error }}
        </div>
      </div>
      <div class="bank-modal-footer">
        <button class="bank-btn bank-btn-outline" (click)="onCancel()">取消</button>
        <button class="bank-btn bank-btn-gold" (click)="onVerify()">验证</button>
      </div>
    </div>
  `
})
export class TwoFADialogComponent {
  code = '';
  error = '';

  constructor(
    public dialogRef: MatDialogRef<TwoFADialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: { target: string; action: string }
  ) {}

  onVerify(): void {
    if (this.code.length !== 6) {
      this.error = '请输入6位验证码';
      return;
    }
    this.dialogRef.close(this.code);
  }

  onCancel(): void {
    this.dialogRef.close(null);
  }
}
