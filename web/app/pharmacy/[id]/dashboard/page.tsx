"use client";

import { useState, useEffect, use } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { CheckIcon, XIcon, AlertTriangleIcon, ClockIcon } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { BASE_URL } from "@/utils/api";

interface MedicationAlert {
  id: number;
  medication_name: string;
  active_substance: string;
  customer_name: string;
  customer_distance_km: number;
  created_at: string;
  expires_at: string;
  already_responded: boolean;
}

interface PharmacyStats {
  alerts_received: number;
  responses_sent: number;
  avg_response_time: number;
  availability_rate: number;
}

export default function PharmacistDashboard({ params }: { params: { id: string } }) {
  const pharmacyId = use(params).id;
  const [alerts, setAlerts] = useState<MedicationAlert[]>([]);
  const [stats, setStats] = useState<PharmacyStats>({
    alerts_received: 0,
    responses_sent: 0,
    avg_response_time: 0,
    availability_rate: 0,
  });
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  useEffect(() => {
    fetchDashboardData();
    
    // Set up real-time updates via Server-Sent Events
    const eventSource = new EventSource(`${BASE_URL}/pharmacy/${pharmacyId}/alerts/stream`);
    
    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === "new_alert") {
        // The backend only pushes a lightweight { type, alert_id } event over
        // SSE (no full alert payload), so refetch the dashboard to get the
        // complete, joined alert data instead of inserting a partial object.
        fetchDashboardData();
        toast({
          title: "New Medication Alert",
          description: "A patient nearby is looking for a medication you may have.",
          duration: 5000,
        });
      } else if (data.type === "alert_expired" || data.type === "alert_completed") {
        setAlerts(prev => prev.filter(alert => alert.id !== data.alert_id));
      }
    };

    return () => {
      eventSource.close();
    };
  }, [pharmacyId, toast]);

  const fetchDashboardData = async () => {
    try {
      const response = await fetch(`${BASE_URL}/pharmacy/${pharmacyId}/dashboard`);
      const data = await response.json();
      
      if (data.success) {
        setAlerts(data.data.pending_alerts || []);
        setStats(data.data.today_stats || stats);
      }
    } catch (error) {
      console.error("Failed to fetch dashboard data:", error);
      toast({
        title: "Error",
        description: "Failed to load dashboard data",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleResponse = async (alertId: number, responseType: string, substituteData?: any) => {
    try {
      const response = await fetch(`${BASE_URL}/alerts/${alertId}/response`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          pharmacy_id: parseInt(pharmacyId),
          response_type: responseType,
          response_time_seconds: Math.floor((Date.now() - new Date(alerts.find(a => a.id === alertId)?.created_at || 0).getTime()) / 1000),
          ...substituteData,
        }),
      });

      const data = await response.json();
      
      if (data.success) {
        // Remove the alert from the list
        setAlerts(prev => prev.filter(alert => alert.id !== alertId));
        
        // Update stats
        setStats(prev => ({
          ...prev,
          responses_sent: prev.responses_sent + 1,
        }));

        toast({
          title: "Response Sent",
          description: `You responded "${responseType}" to the medication request`,
        });
      } else {
        throw new Error(data.error || "Failed to send response");
      }
    } catch (error) {
      console.error("Failed to send response:", error);
      toast({
        title: "Error",
        description: "Failed to send response",
        variant: "destructive",
      });
    }
  };

  const getTimeRemaining = (expiresAt: string) => {
    const remaining = new Date(expiresAt).getTime() - Date.now();
    const minutes = Math.floor(remaining / 60000);
    const seconds = Math.floor((remaining % 60000) / 1000);
    return remaining > 0 ? `${minutes}:${seconds.toString().padStart(2, '0')}` : "Expired";
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="flex items-center justify-center h-64">
          <div className="text-lg">Loading dashboard...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Pharmacist Dashboard</h1>
        <Badge variant="outline" className="text-lg px-3 py-1">
          Pharmacy #{pharmacyId}
        </Badge>
      </div>

      {/* Today's Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Alerts Received</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.alerts_received}</div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Responses Sent</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.responses_sent}</div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Avg Response Time</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.avg_response_time}s</div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Availability Rate</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{Math.round(stats.availability_rate * 100)}%</div>
          </CardContent>
        </Card>
      </div>

      {/* Pending Alerts */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertTriangleIcon className="h-5 w-5" />
            Pending Medication Requests
          </CardTitle>
          <CardDescription>
            Respond quickly to help customers find their medications
          </CardDescription>
        </CardHeader>
        <CardContent>
          {alerts.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No pending alerts. You're all caught up!
            </div>
          ) : (
            <div className="space-y-4">
              {alerts.map((alert) => (
                <div
                  key={alert.id}
                  className="border rounded-lg p-4 space-y-3 bg-card"
                >
                  <div className="flex items-start justify-between">
                    <div className="space-y-1">
                      <h3 className="font-semibold text-lg">{alert.medication_name}</h3>
                      <p className="text-sm text-muted-foreground">
                        Active substance: {alert.active_substance}
                      </p>
                      <p className="text-sm">
                        Customer: {alert.customer_name} • {alert.customer_distance_km.toFixed(1)}km away
                      </p>
                    </div>
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <ClockIcon className="h-4 w-4" />
                      {getTimeRemaining(alert.expires_at)}
                    </div>
                  </div>

                  {!alert.already_responded && (
                    <div className="flex gap-2 pt-2">
                      <Button
                        onClick={() => handleResponse(alert.id, "available")}
                        className="flex items-center gap-2 bg-green-600 hover:bg-green-700"
                      >
                        <CheckIcon className="h-4 w-4" />
                        Available
                      </Button>
                      
                      <Button
                        onClick={() => handleResponse(alert.id, "unavailable")}
                        variant="destructive"
                        className="flex items-center gap-2"
                      >
                        <XIcon className="h-4 w-4" />
                        Not Available
                      </Button>
                      
                      <Button
                        onClick={() => {
                          // TODO: Open substitute selection modal
                          handleResponse(alert.id, "substitute", {
                            substitute_brand: "Generic Alternative",
                            substitute_notes: "Similar medication available",
                          });
                        }}
                        variant="outline"
                        className="flex items-center gap-2"
                      >
                        <AlertTriangleIcon className="h-4 w-4" />
                        Substitute Available
                      </Button>
                    </div>
                  )}

                  {alert.already_responded && (
                    <div className="pt-2">
                      <Badge variant="secondary">Already Responded</Badge>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}