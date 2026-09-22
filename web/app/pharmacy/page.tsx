"use client";

import { useState, useEffect } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  MapPin,
  Loader2,
  Edit3,
  Save,
  X,
  Bell,
  CheckCircle,
  User,
  Pill,
  Trash2,
  Plus,
} from "lucide-react";
import { BASE_URL } from "@/utils/api";
import { useAuth } from "@/contexts/auth-context";
import { toast } from "sonner";

interface Pharmacy {
  id: number;
  name: string;
  address: string;
  phone: string;
  latitude: number;
  longitude: number;
  city: string;
}

interface Alert {
  id: number;
  medication_name: string;
  active_substance: string;
  customer_name: string;
  customer_distance_km: number;
  already_responded: boolean;
  created_at: string;
  expires_at: string;
}

interface StockItem {
  id: number;
  pharmacy_id: number;
  medication_id: number;
  speciality: string;
  quantity: number;
  updated_at: string;
}

interface MedicationSearchResult {
  id: number;
  speciality: string;
  active_substance: string;
  dosage: string;
  form: string;
}

export default function PharmacistDashboard() {
  const { user, isAuthenticated, isLoading } = useAuth();

  // Pharmacy state
  const [pharmacy, setPharmacy] = useState<Pharmacy | null>(null);
  const [loading, setLoading] = useState(true);

  // Editing state
  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState({
    name: "",
    address: "",
    phone: "",
    city: "",
  });
  const [updating, setUpdating] = useState(false);

  // Alerts state
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [alertsLoading, setAlertsLoading] = useState(false);
  const [confirmingAlert, setConfirmingAlert] = useState<number | null>(null);

  // Stock state
  const [stock, setStock] = useState<StockItem[]>([]);
  const [stockLoading, setStockLoading] = useState(false);
  const [stockSearchTerm, setStockSearchTerm] = useState("");
  const [stockSearchResults, setStockSearchResults] = useState<
    MedicationSearchResult[]
  >([]);
  const [stockSelectedMed, setStockSelectedMed] =
    useState<MedicationSearchResult | null>(null);
  const [stockQuantity, setStockQuantity] = useState("10");
  const [addingStock, setAddingStock] = useState(false);

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.managed_pharmacy) {
      fetchPharmacyInfo();
    } else if (!isLoading && isAuthenticated) {
      setLoading(false);
    }
  }, [isAuthenticated, isLoading, user]);

  useEffect(() => {
    if (pharmacy) {
      fetchAlerts();
      fetchStock();
    }
  }, [pharmacy]);

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      searchMedicationsForStock(stockSearchTerm);
    }, 300);
    return () => clearTimeout(timeoutId);
  }, [stockSearchTerm]);

  const searchMedicationsForStock = async (term: string) => {
    if (term.length < 2) {
      setStockSearchResults([]);
      return;
    }
    try {
      const response = await fetch(
        `${BASE_URL}/meds/search?q=${encodeURIComponent(term)}`,
      );
      const data = await response.json();
      if (Array.isArray(data)) {
        setStockSearchResults(data.slice(0, 8));
      }
    } catch (error) {
      console.error("Error searching medications:", error);
    }
  };

  const fetchStock = async () => {
    if (!pharmacy) return;
    setStockLoading(true);
    try {
      const response = await fetch(
        `${BASE_URL}/pharmacies/id/${pharmacy.id}/stock`,
      );
      if (response.ok) {
        const data = await response.json();
        setStock(data.data || []);
      }
    } catch (error) {
      console.error("Error fetching stock:", error);
    } finally {
      setStockLoading(false);
    }
  };

  const addStock = async () => {
    if (!pharmacy || !stockSelectedMed) return;
    const quantity = parseInt(stockQuantity, 10);
    if (isNaN(quantity) || quantity < 0) {
      toast.error("Enter a valid quantity");
      return;
    }
    setAddingStock(true);
    try {
      const response = await fetch(
        `${BASE_URL}/pharmacies/id/${pharmacy.id}/stock`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            medication_id: stockSelectedMed.id,
            quantity,
          }),
        },
      );
      if (response.ok) {
        toast.success(`${stockSelectedMed.speciality} added to your stock`);
        setStockSelectedMed(null);
        setStockSearchTerm("");
        setStockSearchResults([]);
        setStockQuantity("10");
        fetchStock();
      } else {
        throw new Error("Failed to add stock");
      }
    } catch (error) {
      console.error("Error adding stock:", error);
      toast.error("Failed to add medication to stock");
    } finally {
      setAddingStock(false);
    }
  };

  const removeStock = async (medicationId: number) => {
    if (!pharmacy) return;
    try {
      const response = await fetch(
        `${BASE_URL}/pharmacies/id/${pharmacy.id}/stock/${medicationId}`,
        { method: "DELETE" },
      );
      if (response.ok) {
        setStock(stock.filter((item) => item.medication_id !== medicationId));
        toast.success("Removed from stock");
      } else {
        throw new Error("Failed to remove stock");
      }
    } catch (error) {
      console.error("Error removing stock:", error);
      toast.error("Failed to remove medication from stock");
    }
  };

  const fetchPharmacyInfo = async () => {
    setLoading(true);
    try {
      const response = await fetch(
        `${BASE_URL}/pharmacies/id/${user?.managed_pharmacy}`,
      );
      if (response.ok) {
        const data = await response.json();
        setPharmacy(data.pharmacy);
        setEditForm({
          name: data.pharmacy.name,
          address: data.pharmacy.address,
          phone: data.pharmacy.phone,
          city: data.pharmacy.city,
        });
      } else {
        console.error("Failed to fetch pharmacy details");
      }
    } catch (error) {
      console.error("Error fetching pharmacy:", error);
    } finally {
      setLoading(false);
    }
  };

  // Update edit form when pharmacy data changes
  useEffect(() => {
    if (pharmacy) {
      setEditForm({
        name: pharmacy.name,
        address: pharmacy.address,
        phone: pharmacy.phone,
        city: pharmacy.city,
      });
    }
  }, [pharmacy]);

  const fetchAlerts = async () => {
    setAlertsLoading(true);
    try {
      const response = await fetch(
        `${BASE_URL}/pharmacy/${pharmacy?.id}/dashboard`,
      );
      if (response.ok) {
        const data = await response.json();
        setAlerts(data.data?.pending_alerts || []);
      }
    } catch (error) {
      console.error("Error fetching alerts:", error);
    } finally {
      setAlertsLoading(false);
    }
  };

  const handleEditSubmit = async () => {
    if (!pharmacy) return;

    setUpdating(true);
    try {
      const response = await fetch(`${BASE_URL}/pharmacies/id/${pharmacy.id}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(editForm),
      });

      if (response.ok) {
        const updatedPharmacy = await response.json();
        setPharmacy(updatedPharmacy.pharmacy);
        setIsEditing(false);
        toast.success("Pharmacy information updated successfully");
      } else {
        throw new Error("Failed to update pharmacy");
      }
    } catch (error) {
      console.error("Error updating pharmacy:", error);
      toast.error("Failed to update pharmacy information");
    } finally {
      setUpdating(false);
    }
  };

  const confirmAlertAvailability = async (
    alertId: number,
    status: "available" | "unavailable",
  ) => {
    if (!pharmacy) return;
    setConfirmingAlert(alertId);
    try {
      const response = await fetch(`${BASE_URL}/alerts/${alertId}/response`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          pharmacy_id: pharmacy.id,
          response_type: status,
        }),
      });

      if (response.ok) {
        setAlerts(
          alerts.map((alert) =>
            alert.id === alertId
              ? { ...alert, already_responded: true }
              : alert,
          ),
        );
        toast.success(`Alert marked as ${status}`);
      } else {
        throw new Error("Failed to confirm alert");
      }
    } catch (error) {
      console.error("Error confirming alert:", error);
      toast.error("Failed to update alert status");
    } finally {
      setConfirmingAlert(null);
    }
  };

  if (isLoading || loading) {
    return (
      <div className="min-h-screen bg-slate-50 dark:bg-slate-900 flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="h-12 w-12 animate-spin text-teal-600 mx-auto mb-4" />
          <p className="text-slate-600 dark:text-slate-300">
            Loading dashboard...
          </p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-slate-50 dark:bg-slate-900 flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6">
            <div className="text-center">
              <h2 className="text-xl font-semibold mb-2">
                Authentication Required
              </h2>
              <p className="text-slate-600 dark:text-slate-300 mb-4">
                Please sign in to access the pharmacy dashboard
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!pharmacy) {
    return (
      <div className="min-h-screen bg-slate-50 dark:bg-slate-900 flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6">
            <div className="text-center">
              <h2 className="text-xl font-semibold mb-2 text-amber-600">
                No Pharmacy Assigned
              </h2>
              <p className="text-slate-600 dark:text-slate-300 mb-4">
                You don't have a pharmacy assigned to your account. Please
                contact an administrator.
              </p>
              <Button
                onClick={() => window.location.reload()}
                variant="outline"
              >
                Retry
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-2">
            Pharmacy Dashboard
          </h1>
          <p className="text-slate-600 dark:text-slate-300">
            Manage your pharmacy information and respond to medication alerts
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Pharmacy Information */}
          <div className="lg:col-span-2">
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="flex items-center gap-2">
                    <MapPin className="h-5 w-5 text-teal-600" />
                    Pharmacy Information
                  </CardTitle>
                  {!isEditing ? (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setIsEditing(true)}
                    >
                      <Edit3 className="h-4 w-4 mr-2" />
                      Edit
                    </Button>
                  ) : (
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          setIsEditing(false);
                          setEditForm({
                            name: pharmacy.name,
                            address: pharmacy.address,
                            phone: pharmacy.phone,
                            city: pharmacy.city,
                          });
                        }}
                      >
                        <X className="h-4 w-4 mr-2" />
                        Cancel
                      </Button>
                      <Button
                        size="sm"
                        onClick={handleEditSubmit}
                        disabled={updating}
                        className="bg-teal-600 hover:bg-teal-700"
                      >
                        {updating ? (
                          <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        ) : (
                          <Save className="h-4 w-4 mr-2" />
                        )}
                        Save
                      </Button>
                    </div>
                  )}
                </div>
              </CardHeader>
              <CardContent>
                {isEditing ? (
                  <div className="space-y-4">
                    <div>
                      <Label htmlFor="name">Pharmacy Name</Label>
                      <Input
                        id="name"
                        value={editForm.name}
                        onChange={(e) =>
                          setEditForm({ ...editForm, name: e.target.value })
                        }
                        placeholder="Enter pharmacy name"
                      />
                    </div>
                    <div>
                      <Label htmlFor="address">Address</Label>
                      <Input
                        id="address"
                        value={editForm.address}
                        onChange={(e) =>
                          setEditForm({ ...editForm, address: e.target.value })
                        }
                        placeholder="Enter address"
                      />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor="city">City</Label>
                        <Input
                          id="city"
                          value={editForm.city}
                          onChange={(e) =>
                            setEditForm({ ...editForm, city: e.target.value })
                          }
                          placeholder="Enter city"
                        />
                      </div>
                      <div>
                        <Label htmlFor="phone">Phone</Label>
                        <Input
                          id="phone"
                          value={editForm.phone}
                          onChange={(e) =>
                            setEditForm({ ...editForm, phone: e.target.value })
                          }
                          placeholder="Enter phone number"
                        />
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div>
                      <h3 className="font-semibold text-lg flex items-center gap-2">
                        {pharmacy.name}
                      </h3>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                      <div>
                        <p className="text-slate-500 dark:text-slate-400">
                          Address
                        </p>
                        <p className="font-medium">{pharmacy.address}</p>
                      </div>
                      <div>
                        <p className="text-slate-500 dark:text-slate-400">
                          City
                        </p>
                        <p className="font-medium">{pharmacy.city}</p>
                      </div>
                      <div>
                        <p className="text-slate-500 dark:text-slate-400">
                          Phone
                        </p>
                        <p className="font-medium">{pharmacy.phone}</p>
                      </div>
                      <div>
                        <p className="text-slate-500 dark:text-slate-400">
                          Status
                        </p>
                      </div>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Quick Stats */}
          <div>
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Alert Summary</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="text-slate-600 dark:text-slate-300">
                      Awaiting Response
                    </span>
                    <Badge
                      variant="outline"
                      className="bg-yellow-50 text-yellow-700 border-yellow-200"
                    >
                      {alerts.filter((a) => !a.already_responded).length}
                    </Badge>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-slate-600 dark:text-slate-300">
                      Responded
                    </span>
                    <Badge
                      variant="outline"
                      className="bg-green-50 text-green-700 border-green-200"
                    >
                      {alerts.filter((a) => a.already_responded).length}
                    </Badge>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        {/* My Stock */}
        <Card className="mt-6">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Pill className="h-5 w-5 text-teal-600" />
              My Stock ({stock.length})
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-slate-500 dark:text-slate-400">
              Only medications listed here will match when a patient sends an
              alert — add what your pharmacy actually carries.
            </p>

            <div className="border rounded-lg p-4 space-y-3">
              <Label>Add a medication to your stock</Label>
              <Input
                placeholder="Search medication name..."
                value={stockSearchTerm}
                onChange={(e) => {
                  setStockSearchTerm(e.target.value);
                  setStockSelectedMed(null);
                }}
              />
              {stockSearchResults.length > 0 && !stockSelectedMed && (
                <div className="border rounded-md divide-y max-h-56 overflow-y-auto">
                  {stockSearchResults.map((med) => (
                    <div
                      key={med.id}
                      className="p-2 cursor-pointer hover:bg-muted text-sm"
                      onClick={() => {
                        setStockSelectedMed(med);
                        setStockSearchTerm(med.speciality);
                        setStockSearchResults([]);
                      }}
                    >
                      <div className="font-medium">{med.speciality}</div>
                      <div className="text-xs text-slate-500">
                        {med.active_substance} • {med.dosage} • {med.form}
                      </div>
                    </div>
                  ))}
                </div>
              )}
              {stockSelectedMed && (
                <div className="flex items-end gap-2">
                  <div className="w-32">
                    <Label htmlFor="stock-qty">Quantity</Label>
                    <Input
                      id="stock-qty"
                      type="number"
                      min="0"
                      value={stockQuantity}
                      onChange={(e) => setStockQuantity(e.target.value)}
                    />
                  </div>
                  <Button onClick={addStock} disabled={addingStock}>
                    {addingStock ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <>
                        <Plus className="h-4 w-4 mr-1" />
                        Add
                      </>
                    )}
                  </Button>
                </div>
              )}
            </div>

            {stockLoading ? (
              <div className="flex items-center justify-center py-6">
                <Loader2 className="h-5 w-5 animate-spin text-teal-600" />
              </div>
            ) : stock.length > 0 ? (
              <div className="divide-y border rounded-lg">
                {stock.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-center justify-between p-3"
                  >
                    <div>
                      <div className="font-medium">{item.speciality}</div>
                      <div className="text-xs text-slate-500">
                        Quantity: {item.quantity}
                      </div>
                    </div>
                    <Button
                      size="sm"
                      variant="ghost"
                      className="text-red-600 hover:bg-red-50"
                      onClick={() => removeStock(item.medication_id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-slate-400 text-center py-4">
                No stock added yet — search above to add your first
                medication.
              </p>
            )}
          </CardContent>
        </Card>

        {/* Medication Alerts */}
        <Card className="mt-6">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5 text-teal-600" />
              Medication Alerts ({alerts.length})
            </CardTitle>
          </CardHeader>
          <CardContent>
            {alertsLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-teal-600" />
                <span className="ml-2">Loading alerts...</span>
              </div>
            ) : alerts.length > 0 ? (
              <div className="space-y-4">
                {alerts.map((alert) => (
                  <div
                    key={alert.id}
                    className="border rounded-lg p-4 space-y-3"
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          <User className="h-4 w-4 text-slate-500" />
                          <span className="font-medium">
                            {alert.customer_name}
                          </span>
                          <Badge
                            variant={
                              alert.already_responded ? "default" : "secondary"
                            }
                            className={
                              alert.already_responded
                                ? "bg-green-100 text-green-800"
                                : "bg-yellow-100 text-yellow-800"
                            }
                          >
                            {alert.already_responded ? "Responded" : "Pending"}
                          </Badge>
                        </div>
                        <div className="flex items-center gap-2 mb-2">
                          <Pill className="h-4 w-4 text-teal-600" />
                          <span className="font-medium text-teal-700">
                            {alert.medication_name}
                          </span>
                        </div>
                        <p className="text-slate-600 dark:text-slate-300 text-sm">
                          {alert.active_substance} •{" "}
                          {alert.customer_distance_km.toFixed?.(1) ??
                            alert.customer_distance_km}
                          km away
                        </p>
                        <p className="text-xs text-slate-500 mt-1">
                          {new Date(alert.created_at).toLocaleString()}
                        </p>
                      </div>

                      {!alert.already_responded && (
                        <div className="flex gap-2 ml-4">
                          <Button
                            size="sm"
                            variant="outline"
                            className="text-green-600 border-green-200 hover:bg-green-50"
                            onClick={() =>
                              confirmAlertAvailability(alert.id, "available")
                            }
                            disabled={confirmingAlert === alert.id}
                          >
                            {confirmingAlert === alert.id ? (
                              <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                              <>
                                <CheckCircle className="h-4 w-4 mr-1" />
                                Available
                              </>
                            )}
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            className="text-red-600 border-red-200 hover:bg-red-50"
                            onClick={() =>
                              confirmAlertAvailability(alert.id, "unavailable")
                            }
                            disabled={confirmingAlert === alert.id}
                          >
                            {confirmingAlert === alert.id ? (
                              <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                              <>
                                <X className="h-4 w-4 mr-1" />
                                Not Available
                              </>
                            )}
                          </Button>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <Bell className="h-12 w-12 text-slate-300 mx-auto mb-4" />
                <p className="text-slate-500 dark:text-slate-400">
                  No alerts at this time
                </p>
                <p className="text-sm text-slate-400 dark:text-slate-500">
                  Users will be able to create medication availability alerts
                  for your pharmacy
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
