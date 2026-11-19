"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import { useAuth } from "@/contexts/auth-context";
import { useAutoTranslation } from "@/contexts/auto-translation-context";
import { BASE_URL } from "@/utils/api";
import { toast } from "sonner";
import { ChevronDown } from "lucide-react";

type Pickup = { id: number; alert_id: number; quantity: number; status: "awaiting" | "picked_up" | "canceled"; expires_at: string; reminder_stage: number; medication_name: string; pharmacy_name: string; customer_name: string };
const copy = {
 en: { title: "Pickup requests", awaiting: "Awaiting pickup", picked_up: "Picked up", canceled: "Canceled", collect: "Mark as picked up", empty: "No requests", reminder: "Remember to collect your medication.", minute: "Only 1 minute left to collect your medication.", canceledMessage: "Your reservation expired. The reserved stock has been restored.", error: "Could not update pickup requests. Please try again.", retry: "Retry", deadline: "Time remaining", waiting: "Awaiting expiry confirmation", patient: "My medication reservations" },
 fr: { title: "Demandes de retrait", awaiting: "En attente de retrait", picked_up: "Retirées", canceled: "Annulées", collect: "Confirmer le retrait", empty: "Aucune demande", reminder: "Pensez à récupérer votre médicament.", minute: "Il reste une minute pour récupérer votre médicament.", canceledMessage: "Votre réservation a expiré. La quantité réservée a été remise en stock.", error: "Impossible de mettre à jour les demandes. Réessayez.", retry: "Réessayer", deadline: "Temps restant", waiting: "En attente de confirmation d’expiration", patient: "Mes réservations de médicaments" },
 ar: { title: "طلبات الاستلام", awaiting: "في انتظار الاستلام", picked_up: "تم الاستلام", canceled: "ملغاة", collect: "تأكيد الاستلام", empty: "لا توجد طلبات", reminder: "تذكير باستلام دوائك.", minute: "بقيت دقيقة واحدة لاستلام دوائك.", canceledMessage: "انتهت صلاحية الحجز وأعيدت الكمية المحجوزة إلى المخزون.", error: "تعذر تحديث الطلبات. يرجى المحاولة مجددًا.", retry: "إعادة المحاولة", deadline: "الوقت المتبقي", waiting: "في انتظار تأكيد انتهاء الحجز", patient: "حجوزاتي للأدوية" },
};
export function PickupRequests({ pharmacyId }: { pharmacyId?: number }) {
 const { user } = useAuth();
 const { language } = useAutoTranslation(); const t = copy[language];
 const [items, setItems] = useState<Pickup[]>([]); const [error, setError] = useState(false);
 const [now, setNow] = useState(Date.now()); const [busy, setBusy] = useState<number | null>(null);
 const [expandedStatus, setExpandedStatus] = useState<Pickup["status"] | null>(null);
 const seen = useRef(new Set<string>());
 const listsRef = useRef<HTMLDivElement>(null);
 useEffect(() => {
  if (!expandedStatus) return;
  const openList = () => listsRef.current?.querySelector<HTMLDetailsElement>("details[open]");
  const closeOutside = (event: Event) => {
   if (event.target instanceof Node && !openList()?.contains(event.target)) setExpandedStatus(null);
  };
  const closeOnEscape = (event: KeyboardEvent) => {
   if (event.key !== "Escape") return;
   openList()?.querySelector("summary")?.focus();
   setExpandedStatus(null);
  };
  document.addEventListener("pointerdown", closeOutside, true);
  document.addEventListener("focusin", closeOutside);
  document.addEventListener("keydown", closeOnEscape);
  return () => {
   document.removeEventListener("pointerdown", closeOutside, true);
   document.removeEventListener("focusin", closeOutside);
   document.removeEventListener("keydown", closeOnEscape);
  };
 }, [expandedStatus]);
 const load = useCallback(async (signal?: AbortSignal) => {
  if (!user) return;
  try {
   const path = pharmacyId ? `pharmacy/${pharmacyId}` : "my";
   const res = await fetch(`${BASE_URL}/alerts/pickups/${path}`, { headers: { Authorization: `Bearer ${sessionStorage.getItem("token") || ""}` }, cache: "no-store", signal });
   if (!res.ok) throw new Error(); const data = await res.json();
   if (signal?.aborted) return; setItems(data.data || []); setError(false);
  } catch { if (!signal?.aborted) setError(true); }
 }, [user?.id, pharmacyId]);
 useEffect(() => {
  setItems([]); const controller = new AbortController();
  const refresh = () => { void load(controller.signal); };
  refresh(); const interval = setInterval(refresh, 2000);
  window.addEventListener("qayen-pickups-changed", refresh);
  return () => { controller.abort(); clearInterval(interval); window.removeEventListener("qayen-pickups-changed", refresh); };
 }, [load]);
 useEffect(() => { const timer = setInterval(() => setNow(Date.now()), 1000); return () => clearInterval(timer); }, []);
 useEffect(() => {
  if (!user || pharmacyId) return;
  for (const item of items) {
   const stage = item.status === "canceled" ? "canceled" : item.status === "awaiting" && item.reminder_stage > 0 ? String(item.reminder_stage) : "";
   if (!stage) continue; const key = `qayen-pickup:${user.id}:${item.id}:${stage}`;
   let notified = seen.current.has(key); try { notified ||= localStorage.getItem(key) === "shown"; } catch {}
   if (notified) continue; seen.current.add(key); try { localStorage.setItem(key, "shown"); } catch {}
   toast(stage === "canceled" ? t.canceledMessage : stage === "2" ? t.minute : t.reminder, { description: `${item.medication_name} · ${item.pharmacy_name}`, duration: 12000 });
  }
 }, [items, pharmacyId, user?.id, t]);
 async function collect(id: number) {
  setBusy(id);
  try { const res = await fetch(`${BASE_URL}/alerts/pickups/${id}/collected`, { method: "POST", headers: { Authorization: `Bearer ${sessionStorage.getItem("token") || ""}` } }); if (!res.ok) throw new Error(); await load(); window.dispatchEvent(new Event("qayen-pickups-changed")); }
  catch { toast.error(t.error); void load(); } finally { setBusy(null); }
 }
 if (!user || (!pharmacyId && !items.length && !error)) return null;
 return <section translate="no" dir={language === "ar" ? "rtl" : undefined} className="container mx-auto my-3 rounded-xl border bg-background p-3" aria-label={t.title}>
  <h2 className="mb-2 text-base font-semibold">{pharmacyId ? t.title : t.patient}</h2>
  {error && <p role="alert" className="mb-3 text-red-600">{t.error} <button className="underline" onClick={() => void load()}>{t.retry}</button></p>}
  <div ref={listsRef} className="grid items-start gap-2 md:grid-cols-3">{(["awaiting", "picked_up", "canceled"] as const).map(status => {
   const requests = items.filter(i => i.status === status);
   return <details key={status} open={expandedStatus === status} className="group relative min-w-0 rounded-lg border open:z-40">
   <summary onClick={(event) => { event.preventDefault(); setExpandedStatus(current => current === status ? null : status); }} className="flex cursor-pointer list-none items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-teal-600 [&::-webkit-details-marker]:hidden">
    <span>{t[status]}</span>
    <span className="rounded-full bg-muted px-2 py-0.5 text-xs tabular-nums">{requests.length}</span>
    <ChevronDown aria-hidden="true" className="ms-auto h-4 w-4 shrink-0 transition-transform group-open:rotate-180" />
   </summary>
   <div className="absolute inset-x-0 top-full z-50 mt-2 max-h-64 space-y-2 overflow-y-auto rounded-lg border bg-background p-2 shadow-lg">
   {requests.length === 0 && <p className="px-1 py-2 text-sm text-muted-foreground">{t.empty}</p>}
   {requests.map(item => { const seconds = Math.max(0, Math.ceil((Date.parse(item.expires_at) - now) / 1000)); return <article key={item.id} className="rounded border bg-muted/30 p-3 text-sm">
    <p className="font-medium">{item.medication_name} × {item.quantity}</p><p>{pharmacyId ? item.customer_name : item.pharmacy_name}</p>
    {status === "awaiting" && <><p>{seconds > 0 ? `${t.deadline}: ${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")}` : t.waiting}</p>
     {!pharmacyId && item.reminder_stage > 0 && <p className="mt-2 font-medium text-amber-700" role="status">{item.reminder_stage === 2 ? t.minute : t.reminder}</p>}
     {pharmacyId && <button disabled={seconds === 0 || busy === item.id} onClick={() => void collect(item.id)} className="mt-2 rounded bg-teal-700 px-3 py-2 text-white disabled:opacity-50">{t.collect}</button>}</>}
   </article>; })}
   </div>
  </details>; })}</div>
 </section>;
}
// Patient reminders remain available when navigating to another page.
export function PatientPickupRequests() {
 const { user } = useAuth();
 return user && !user.managed_pharmacy ? <PickupRequests key={user.id} /> : null;
}
