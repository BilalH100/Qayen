"use client";
import { BASE_URL } from "@/utils/api";
import { ChevronLeft, ChevronRight, Edit3, Trash2, UserCog } from "lucide-react";
import React from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Label } from "@/components/ui/label";
import { AdminProtectedRoute } from "@/components/admin-protected-route";
import { toast } from "sonner";
import { useToast } from "@/hooks/use-toast";

type User = {
  id: number;
  name: string;
  email: string;
  role: string;
  phone: string;
};

type UserRole = "regular" | "pharmacist" | "admin";

function AdminDashboardContent() {
  const [loading, setLoading] = React.useState(false);
  const [limit, setLimit] = React.useState(10);
  const [users, setUsers] = React.useState<User[]>([]);
  const [page, setPage] = React.useState(1);
  const [totalUsers, setTotalUsers] = React.useState(0);
  const [regularUsers, setRegularUsers] = React.useState(0);
  const [pharmacists, setPharmacists] = React.useState(0);
  
  // Modal state
  const [isModalOpen, setIsModalOpen] = React.useState(false);
  const [selectedUser, setSelectedUser] = React.useState<User | null>(null);
  const [selectedRole, setSelectedRole] = React.useState<UserRole>("regular");
  const [updating, setUpdating] = React.useState(false);
  
  const [isDeleteModalOpen, setIsDeleteModalOpen] = React.useState(false);
  const [userToDelete, setUserToDelete] = React.useState<User | null>(null);
  const [deleting, setDeleting] = React.useState(false);

  const { toast } = useToast();
  async function fetchUsers() {
    setLoading(true);
    try {
      const res = await fetch(
        `${BASE_URL}/users/list?page=${page}&limit=${limit}`,
      );
      const data = await res.json();
      const usersList = data.users || [];
      setUsers(usersList);
      setTotalUsers(data.total || 0);
      
      // Count users by role
      const regularCount = usersList.filter((user: User) => 
        user.role === 'user' || user.role === 'regular' || user.role === 'client'
      ).length;
      const pharmacistCount = usersList.filter((user: User) => 
        user.role === 'pharmacist' || user.role === 'pharmacy'
      ).length;
      
      setRegularUsers(regularCount);
      setPharmacists(pharmacistCount);
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  }

  React.useEffect(() => {
    fetchUsers();
  }, [page, limit]);

  const totalPages = Math.ceil(totalUsers / limit);

  const openEditModal = (user: User) => {
    setSelectedUser(user);
    setSelectedRole(user.role as UserRole);
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setSelectedUser(null);
    setSelectedRole("regular");
    setUpdating(false);
  };

  const updateUserRole = async () => {
    if (!selectedUser) return;
    
    setUpdating(true);
    try {
      const res = await fetch(`${BASE_URL}/users/${selectedUser.id}/role`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ role: selectedRole }),
      });

      if (res.ok) {
        setUsers(prevUsers => 
          prevUsers.map(user => 
            user.id === selectedUser.id 
              ? { ...user, role: selectedRole }
              : user
          )
        );
        closeModal();
        fetchUsers();
      } else {
        console.error('Failed to update user role');
        toast({
          title: "Role update failed",
          description: "Please try again",
          variant: "destructive",
        });
      }
    } catch (error) {
      console.error('Error updating user role:', error);
    } finally {
      setUpdating(false);
    }
  };

  const openDeleteModal = (user: User) => {
    setUserToDelete(user);
    setIsDeleteModalOpen(true);
  };

  const closeDeleteModal = () => {
    setIsDeleteModalOpen(false);
    setUserToDelete(null);
    setDeleting(false);
  };

  const deleteUser = async () => {
    if (!userToDelete) return;
    console.log("tod delete", userToDelete.id)
    setDeleting(true);
    try {
      const res = await fetch(`${BASE_URL}/users/id/${userToDelete.id}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (res.ok) {
        setUsers(prevUsers => 
          prevUsers.filter(user => user.id !== userToDelete.id)
        );
        closeDeleteModal();
        fetchUsers();
      } else {
        console.error('Failed to delete user');
        toast({
          title: "User deletion failed",
          description: "Please try again",
          variant: "destructive",
        });
      }
    } catch (error) {
      console.error('Error deleting user:', error);
    } finally {
      setDeleting(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
              Admin Dashboard
            </h1>
          </div>
          <p className="text-lg text-slate-600 dark:text-slate-300">
            Manage users and their roles across the platform
          </p>
        </div>

        {/* Stats Cards */}
        <div className="grid gap-6 md:grid-cols-3 mb-8">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Users</CardTitle>
              <div className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{totalUsers}</div>
              <p className="text-xs text-muted-foreground">
                Registered users
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Regular Users</CardTitle>
              <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{regularUsers}</div>
              <p className="text-xs text-muted-foreground">
                Standard accounts
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Pharmacists</CardTitle>
              <div className="w-2 h-2 bg-teal-500 rounded-full"></div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{pharmacists}</div>
              <p className="text-xs text-muted-foreground">
                Healthcare providers
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Users Table */}
        <Card>
          <CardHeader>
            <CardTitle>Users Management</CardTitle>
            <CardDescription>
              View and manage all registered users
            </CardDescription>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="flex items-center justify-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-teal-600"></div>
                <span className="ml-3 text-slate-600 dark:text-slate-300">Loading users...</span>
              </div>
            ) : (
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Email</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead>Phone</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {users.length > 0 ? (
                      users.map((user) => (
                        <TableRow key={user.id}>
                          <TableCell className="font-medium">{user.name}</TableCell>
                          <TableCell>{user.email}</TableCell>
                          <TableCell>
                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                              user.role === 'admin' 
                                ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
                                : 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
                            }`}>
                              {user.role}
                            </span>
                          </TableCell>
                          <TableCell>{user.phone}</TableCell>
                          <TableCell className="text-right">
                            <div className="flex justify-end gap-2">
                              <Dialog open={isModalOpen && selectedUser?.id === user.id} onOpenChange={(open) => {
                                if (!open) closeModal();
                              }}>
                                <DialogTrigger asChild>
                                  <Button variant="outline" size="sm" onClick={() => openEditModal(user)}>
                                    <Edit3 className="h-4 w-4 mr-1" />
                                    Edit Role
                                  </Button>
                                </DialogTrigger>
                              </Dialog>
                              
                              <Dialog open={isDeleteModalOpen && userToDelete?.id === user.id} onOpenChange={(open) => {
                                if (!open) closeDeleteModal();
                              }}>
                                <DialogTrigger asChild>
                                  <Button 
                                    variant="outline" 
                                    size="sm" 
                                    className="text-red-600 border-red-200 hover:bg-red-50 hover:border-red-300 dark:text-red-400 dark:border-red-800 dark:hover:bg-red-950"
                                    onClick={() => openDeleteModal(user)}
                                  >
                                    <Trash2 className="h-4 w-4 mr-1" />
                                    Delete
                                  </Button>
                                </DialogTrigger>
                              </Dialog>
                            </div>
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={5} className="text-center py-8 text-slate-500 dark:text-slate-400">
                          No users found
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
            )}

            {/* Pagination */}
            <div className="flex items-center justify-between mt-6">
              <p className="text-sm text-slate-600 dark:text-slate-300">
                Showing {((page - 1) * limit) + 1} to {Math.min(page * limit, totalUsers)} of {totalUsers} users
              </p>
              
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page === 1}
                  onClick={() => setPage(page - 1)}
                >
                  <ChevronLeft className="h-4 w-4 mr-1" />
                  Previous
                </Button>

                <Button
                  variant="outline"
                  size="sm"
                  disabled={page === totalPages}
                  onClick={() => setPage(page + 1)}
                >
                  Next
                  <ChevronRight className="h-4 w-4 ml-1" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Edit Role Modal */}
        <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
          <DialogContent className="sm:max-w-[425px]">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <UserCog className="h-5 w-5 text-teal-600 dark:text-teal-400" />
                Edit User Role
              </DialogTitle>
              <DialogDescription>
                Update the role for {selectedUser?.name}. This will change their access permissions.
              </DialogDescription>
            </DialogHeader>
            
            <div className="grid gap-4 py-4">
              <div className="space-y-3">
                <Label htmlFor="role-selection" className="text-sm font-medium">
                  Select Role
                </Label>
                <RadioGroup
                  value={selectedRole}
                  onValueChange={(value) => setSelectedRole(value as UserRole)}
                  className="grid gap-3"
                >
                  <div className="flex items-center space-x-3 rounded-lg border p-3 hover:bg-slate-50 dark:hover:bg-slate-800">
                    <RadioGroupItem value="regular" id="regular" />
                    <div className="space-y-1 flex-1">
                      <Label htmlFor="regular" className="cursor-pointer">
                        <div className="font-medium">Regular User</div>
                        <div className="text-sm text-muted-foreground">
                          Standard user with basic access permissions
                        </div>
                      </Label>
                    </div>
                    <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                  </div>
                  
                  <div className="flex items-center space-x-3 rounded-lg border p-3 hover:bg-slate-50 dark:hover:bg-slate-800">
                    <RadioGroupItem value="pharmacist" id="pharmacist" />
                    <div className="space-y-1 flex-1">
                      <Label htmlFor="pharmacist" className="cursor-pointer">
                        <div className="font-medium">Pharmacist</div>
                        <div className="text-sm text-muted-foreground">
                          Healthcare provider with pharmacy management access
                        </div>
                      </Label>
                    </div>
                    <div className="w-2 h-2 bg-teal-500 rounded-full"></div>
                  </div>
                  
                  <div className="flex items-center space-x-3 rounded-lg border p-3 hover:bg-slate-50 dark:hover:bg-slate-800">
                    <RadioGroupItem value="admin" id="admin" />
                    <div className="space-y-1 flex-1">
                      <Label htmlFor="admin" className="cursor-pointer">
                        <div className="font-medium">Administrator</div>
                        <div className="text-sm text-muted-foreground">
                          Full system access with user management capabilities
                        </div>
                      </Label>
                    </div>
                    <div className="w-2 h-2 bg-red-500 rounded-full"></div>
                  </div>
                </RadioGroup>
              </div>
            </div>
            
            <DialogFooter>
              <Button variant="outline" onClick={closeModal} disabled={updating}>
                Cancel
              </Button>
              <Button onClick={updateUserRole} disabled={updating}>
                {updating ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                    Updating...
                  </>
                ) : (
                  'Update Role'
                )}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Delete Confirmation Modal */}
        <Dialog open={isDeleteModalOpen} onOpenChange={setIsDeleteModalOpen}>
          <DialogContent className="sm:max-w-[425px]">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2 text-red-600 dark:text-red-400">
                <Trash2 className="h-5 w-5" />
                Delete User
              </DialogTitle>
              <DialogDescription>
                Are you sure you want to delete <strong>{userToDelete?.name}</strong>? This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            
            <div className="py-4">
              <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-4">
                <div className="flex items-start gap-3">
                  <div className="flex-shrink-0">
                    <div className="w-2 h-2 bg-red-500 rounded-full mt-2"></div>
                  </div>
                  <div className="space-y-2">
                    <p className="font-medium text-red-800 dark:text-red-200">
                      This will permanently delete:
                    </p>
                    <ul className="text-sm text-red-700 dark:text-red-300 space-y-1">
                      <li>• User account and profile information</li>
                      <li>• All associated data and preferences</li>
                      <li>• Access permissions and role assignments</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
            
            <DialogFooter className="gap-2">
              <Button variant="outline" onClick={closeDeleteModal} disabled={deleting}>
                Cancel
              </Button>
              <Button 
                variant="destructive" 
                onClick={deleteUser} 
                disabled={deleting}
                className="bg-red-600 hover:bg-red-700"
              >
                {deleting ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                    Deleting...
                  </>
                ) : (
                  <>
                    <Trash2 className="h-4 w-4 mr-2" />
                    Delete User
                  </>
                )}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
}

export default function AdminDashboard() {
  return (
    <AdminProtectedRoute>
      <AdminDashboardContent />
    </AdminProtectedRoute>
  );
}
