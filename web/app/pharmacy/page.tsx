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
  user_id: number;
  user_name: string;
  medication_name: string;
  medication_id: number;
  pharmacy_id: number;
  message: string;
  status: string; // 'pending', 'available', 'unavailable'
  created_at: string;
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

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.managed_pharmacy_id) {
      fetchPharmacyInfo();
    } else if (!isLoading && isAuthenticated) {
      setLoading(false);
    }
  }, [isAuthenticated, isLoading, user]);

  useEffect(() => {
    if (pharmacy) {
      fetchAlerts();
    }
  }, [pharmacy]);

  const fetchPharmacyInfo = async () => {
    setLoading(true);
    try {
      const response = await fetch(
        `${BASE_URL}/pharmacies/id/${user?.managed_pharmacy_id}`,
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
        `${BASE_URL}/pharmacies/${pharmacy?.id}/alerts`,
      );
      if (response.ok) {
        const data = await response.json();
        setAlerts(data.alerts || []);
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
      const response = await fetch(`${BASE_URL}/pharmacies/${pharmacy.id}`, {
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
    setConfirmingAlert(alertId);
    try {
      const response = await fetch(`${BASE_URL}/alerts/${alertId}/confirm`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ status }),
      });

      if (response.ok) {
        setAlerts(
          alerts.map((alert) =>
            alert.id === alertId ? { ...alert, status } : alert,
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
                      Pending
                    </span>
                    <Badge
                      variant="outline"
                      className="bg-yellow-50 text-yellow-700 border-yellow-200"
                    >
                      {alerts.filter((a) => a.status === "pending").length}
                    </Badge>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-slate-600 dark:text-slate-300">
                      Available
                    </span>
                    <Badge
                      variant="outline"
                      className="bg-green-50 text-green-700 border-green-200"
                    >
                      {alerts.filter((a) => a.status === "available").length}
                    </Badge>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-slate-600 dark:text-slate-300">
                      Unavailable
                    </span>
                    <Badge
                      variant="outline"
                      className="bg-red-50 text-red-700 border-red-200"
                    >
                      {alerts.filter((a) => a.status === "unavailable").length}
                    </Badge>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

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
                          <span className="font-medium">{alert.user_name}</span>
                          <Badge
                            variant={
                              alert.status === "pending"
                                ? "secondary"
                                : alert.status === "available"
                                  ? "default"
                                  : "destructive"
                            }
                            className={
                              alert.status === "pending"
                                ? "bg-yellow-100 text-yellow-800"
                                : alert.status === "available"
                                  ? "bg-green-100 text-green-800"
                                  : "bg-red-100 text-red-800"
                            }
                          >
                            {alert.status.charAt(0).toUpperCase() +
                              alert.status.slice(1)}
                          </Badge>
                        </div>
                        <div className="flex items-center gap-2 mb-2">
                          <Pill className="h-4 w-4 text-teal-600" />
                          <span className="font-medium text-teal-700">
                            {alert.medication_name}
                          </span>
                        </div>
                        <p className="text-slate-600 dark:text-slate-300 text-sm">
                          {alert.message}
                        </p>
                        <p className="text-xs text-slate-500 mt-1">
                          {new Date(alert.created_at).toLocaleString()}
                        </p>
                      </div>

                      {alert.status === "pending" && (
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
