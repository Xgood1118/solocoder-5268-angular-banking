import { Component, Inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';

@Component({
  selector: 'app-confirm-dialog',
  template: `
    <div class="bank-modal">
      <div class="bank-modal-header">
        <span>{{ data.title }}</span>
        <button class="bank-modal-close" (click)="onNo()">&times;</button>
      </div>
      <div class="bank-modal-body">
        <p>{{ data.message }}</p>
      </div>
      <div class="bank-modal-footer">
        <button class="bank-btn bank-btn-outline" (click)="onNo()">取消</button>
        <button class="bank-btn bank-btn-primary" (click)="onYes()">确认</button>
      </div>
    </div>
  `
})
export class ConfirmDialogComponent {
  constructor(
    public dialogRef: MatDialogRef<ConfirmDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: { title: string; message: string }
  ) {}

  onYes(): void {
    this.dialogRef.close(true);
  }

  onNo(): void {
    this.dialogRef.close(false);
  }
}
