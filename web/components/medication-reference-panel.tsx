export interface MedicationReference {
  code: string;
  public_price: string | null;
  reimbursement_base: string | null;
  reimbursement_rate: string;
  source_file: string;
  currency: string;
}

export function MedicationReferencePanel({ references }: { references: MedicationReference[] }) {
  if (references.length === 0) return null;

  return (
    <details className="mt-6 rounded-lg border border-slate-200 p-4 dark:border-slate-700">
      <summary className="cursor-pointer font-medium">CNOPS reference information</summary>
      <p className="mt-3 text-sm text-slate-500 dark:text-slate-400">
        These reference prices come from the supplied CNOPS file and may differ from the AMMPS catalog. Their effective date needs confirmation.
      </p>
      {references.map((reference) => (
        <div key={reference.code} className="mt-4 space-y-1 border-t pt-3 text-sm">
          <p><span className="font-medium">Product code:</span> <span translate="no">{reference.code}</span></p>
          {reference.public_price !== null && <p><span className="font-medium">CNOPS public price:</span> {reference.public_price} {reference.currency}</p>}
          {reference.reimbursement_base !== null && <p><span className="font-medium">Reimbursement base:</span> {reference.reimbursement_base} {reference.currency}</p>}
          {reference.reimbursement_rate && <p><span className="font-medium">Reimbursement rate:</span> {reference.reimbursement_rate}</p>}
          <p className="break-all text-xs text-slate-500"><span>Source file:</span> <span translate="no">{reference.source_file}</span></p>
        </div>
      ))}
    </details>
  );
}
