package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UserType int

const (
 	Patient UserType = iota
    Donor
)

type Hospital struct {
	Total            int                   `json:"total"`
	TotalPatients    int                   `json:"total_patients"`
	TotalDonors      int                   `json:"total_donors"`
	Patients         map[int]User          `json:"patients"`
	Donors           map[int]User          `json:"donors"`
	SecretCodesToIds map[int]UserProtected `json:"secret_codes"`
	IdsToSecretCodes map[int]int `json:"ids_to_secret_codes"`
}

type User struct {
	Id                int      `json:"id"`
	Name              string   `json:"name"`
	Address           string   `json:"address"`
	PhoneNo           string   `json:"phone_no"`
	Type              UserType `json:"type"`
	DiseaseDesc       string   `json:"disease_desc,omitempty"`
	RequestedUserIds  []int    `json:"requested_user_ids"`
	PendingUserIds    []int    `json:"pending_user_ids"`
	ConnectedUsersIds []int    `json:"connected_users_ids"`
}

type UserProtected struct {
	Id   int 	  `json:"id,omitempty"`
	Type UserType `json:"type,omitempty"`
}

//declaring store struct
type usersHandler struct{
	sync.Mutex
	store Hospital
}

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

//creating store
func newUsersHandler()*usersHandler{
	return &usersHandler{
		Mutex: sync.Mutex{},
		store: Hospital{
			Total: 0, 
			TotalPatients: 0, 
			TotalDonors: 0, 
			Patients: map[int]User{}, 
			Donors: map[int]User{}, 
			SecretCodesToIds: map[int]UserProtected{},
			IdsToSecretCodes: map[int]int{},   //map[secret_code] = userId;
		},
	}
}

//helper func
func removeDuplicates(arr []int) []int {
    occured := map[int]bool{}
    result := []int{}
    for e := range arr {
        if occured[arr[e]] != true {
            occured[arr[e]] = true
            result = append(result, arr[e])
        }
    }
  
    return result
}

func find(a []int, x int) int {
	for i, n := range a {
			if x == n {
					return i
			}
	}
	return -1
}

func removeElementByIndex(s []int, index int) []int {
	return append(s[:index], s[index+1:]...)
}

//api routes func
// /users/
func (h *usersHandler) users(w http.ResponseWriter, r *http.Request){
	parts := strings.Split(r.URL.String(), "/");
	path := parts[2];
	
	switch r.Method{
		// users/{path}
		case "GET":
				switch path{
				case "login":
					if(len(parts)<4){
						w.WriteHeader(http.StatusBadRequest);
						w.Write([]byte(fmt.Sprintf("err:  check if secret code given with login/{:secret_code}")))
						return;
					}
					h.login(w,r,parts[3])
					return;		
				
				case "donors":
					h.getAll(w,r,"d");
				
				case "patients":
					h.getAll(w,r,"p");
					return;
				
				default:
					w.WriteHeader(http.StatusBadRequest);
					w.Write([]byte(fmt.Sprintf("err:  check request url path")))
					return;
				}
			return;
		
		case "POST":
				switch path{
				case "signup":
					h.signup(w,r);
					return;
					
				default:
					w.WriteHeader(http.StatusBadRequest);
					w.Write([]byte(fmt.Sprintf("err:  check request url path")))
					return; 
				}

		default:
			w.WriteHeader(http.StatusBadRequest);
			w.Write([]byte(fmt.Sprintf("err:  check request url path")))
			return;
	}

}

// /user/{id}
func (h *usersHandler) user(w http.ResponseWriter, r *http.Request){
	parts := strings.Split(r.URL.String(), "/");
	partsLen := len(parts);

	if partsLen < 3 {
		w.WriteHeader(http.StatusBadRequest);
		w.Write([]byte(fmt.Sprintf("err:  check request url path")))
		return;
	} 

	switch partsLen{
	case 3:
		switch r.Method{
			// /user/{id}
		case "GET":
			h.getUser(w,r,parts[2])
			return

		case "DELETE":
			h.deleteUser(w,r,parts[2]);
			return;
		
		case "PUT":
			h.updateUserContact(w,r,parts[2]);
			return

		default:
			w.WriteHeader(http.StatusBadRequest);
			w.Write([]byte(fmt.Sprintf("err:  check request url path")))
			return; 
		}
			
	case 5:
		if(parts[3] != "request"){
			w.WriteHeader(http.StatusBadRequest);
			w.Write([]byte(fmt.Sprintf("err:  check request url path")))
			return;
		}

		switch r.Method{
			// /user/{id}/request/{id}
	
		case "ACCEPT":
			h.acceptRequest(w,r,parts[2], parts[4]);
			return
		 
		case "SEND":
			h.sendRequest(w,r, parts[2], parts[4]);
			return

		case "DELETE":
			h.cancelRequest(w,r, parts[2], parts[4])
			return

		case "PURGE":
			h.cancelConnection(w,r, parts[2], parts[4])
			return

		default:
			w.WriteHeader(http.StatusBadRequest);
			w.Write([]byte(fmt.Sprintf("err:  check request url path")))
			return; 
		}

	default:
		w.WriteHeader(http.StatusBadRequest);
		w.Write([]byte(fmt.Sprintf("err:  check request url path")))
		return; 
	}
}

//api actions
func (h *usersHandler) login(w http.ResponseWriter, r *http.Request, secretCode string){
	code, err := strconv.Atoi(secretCode);
	fmt.Println(code);

	if(err != nil){
		w.WriteHeader(http.StatusNotFound);
		w.Write([]byte(fmt.Sprintf("err: Invalid Secret Code. Check secret code value")))
		return;
	}

	h.Lock()
		userDetails,status := h.store.SecretCodesToIds[code];
	h.Unlock()

	fmt.Println(userDetails);
	fmt.Println(rand.Int());
	fmt.Println(rand.Int());

	if(!status){
		w.WriteHeader(http.StatusNotFound);
		w.Write([]byte(fmt.Sprintf("err: No user details found. Check secret code value")))
		return;
	}

	switch(userDetails.Type){
		case Patient:
			h.Lock()
				patient, status := h.store.Patients[userDetails.Id];
			h.Unlock()

			fmt.Println(status);

			if(!status){
				w.WriteHeader(http.StatusInternalServerError);
				w.Write([]byte(fmt.Sprintf("err:  User Not Found. Check secret code value")))
				return;
			}
			
			fmt.Println(patient);

			jsonBytes, e := json.Marshal(patient)
			if e !=nil{
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(e.Error()))
				return
			}

			w.Header().Add("content-type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write((jsonBytes))
			return;
		
		case Donor:
			h.Lock()
				donor, status := h.store.Donors[userDetails.Id];
			h.Unlock()

			if(!status){
				w.WriteHeader(http.StatusInternalServerError);
				w.Write([]byte(fmt.Sprintf("err:  User Not Found. Check secret code value")))
				return;
			}
			
			fmt.Println(donor);

			jsonBytes, e := json.Marshal(donor)
			if e !=nil{
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(e.Error()))
				return
			}

			w.Header().Add("content-type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write((jsonBytes))
			return;
		}
}

//get all donors or patients
func (h *usersHandler) getAll(w http.ResponseWriter, r *http.Request,t string){
	switch(t){
	case "d":
		donors := make([]User, h.store.TotalDonors);
		h.Lock()
		i := 0
		for _, donor := range h.store.Donors{
			donors[i] = donor
			i++
		}
		h.Unlock()

		jsonBytes, err := json.Marshal(donors)
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write((jsonBytes))
		return;
	
	case "p":
		patients := make([]User, h.store.TotalPatients);
		h.Lock()
		i := 0
		for _, patient := range h.store.Patients{
			patients[i] = patient
			i++
		}
		h.Unlock()

		jsonBytes, err := json.Marshal(patients)
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write((jsonBytes))
		return;
	}
}

func (h *usersHandler) signup(w http.ResponseWriter, r *http.Request){
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close();
	
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
		}

	ct := r.Header.Get("content-type")
	
	if ct != "application/json"{
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("err:  required content-type: application/json but got '%s'", ct)))
	}

	var user User
	e := json.Unmarshal(bodyBytes, &user)

	if e != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return 
	}

	if user.Name == ""{
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf("err:  required name but got empty string")))
		return
	}

	if user.Address == ""{
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf("err:  required address but got empty string")))
		return
	}


	if user.PhoneNo == ""{
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf("err:  required phone but got empty string")))
		return
	}

	if user.Type != 0 && user.Type != 1{
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf("err: enter valid user type. \n 0: Patient \n 1: Donor")))
		return
	}

	h.Lock()
	defer h.Unlock();

	//adding to store
	
	h.store.Total += 1;

	if(user.Type == Donor){
		h.store.TotalDonors += 1;
		user.Id = h.store.Total;
		fmt.Println(user);
		h.store.Donors[user.Id] = user;
	}

	if(user.Type == Patient){
		h.store.TotalPatients += 1;
		user.Id = h.store.Total;
		fmt.Println(user);
		h.store.Patients[user.Id] = user;
	}
		
	fmt.Println("user stored");
	
	secretCode := rand.Int();
	h.store.SecretCodesToIds[secretCode] = UserProtected{
		Id: user.Id,
		Type: user.Type,
	};

	h.store.IdsToSecretCodes[user.Id] = secretCode;
	
	type Data struct {
		UserInfo       User   `json:"user_data,omitempty"`
		UserSecretCode int `json:"user_secret_code,omitempty"`
	}

	userData := Data{
		UserInfo: user,
		UserSecretCode: secretCode,
	}

	//returning to server
	jsonBytes, err := json.Marshal(userData)
	jsonBytes = append(jsonBytes, )
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		h.Unlock()
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write((jsonBytes))
	return;
}

//getUser
func (h *usersHandler) getUser(w http.ResponseWriter, r *http.Request,t string){

	fmt.Println("\n get user started ");

	userId, err := strconv.Atoi(t);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid User id. Check Input UserId")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	userSecretCode, ok := h.store.IdsToSecretCodes[userId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input UserId")))
		return
	}
	
	h.Lock()
	userConfig, ok := h.store.SecretCodesToIds[userSecretCode];
	h.Unlock()
	
	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: UserId and Secret Code Mismatched. Contact Admin")))
		return
	}

	if userConfig.Type == Patient{
		h.Lock()
		patient := h.store.Patients[userConfig.Id];
		h.Unlock()

		jsonBytes, err := json.Marshal(patient)
	
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write((jsonBytes))
		return;
	}

	if userConfig.Type == Donor{
		h.Lock()
			donor := h.store.Donors[userConfig.Id];
		h.Unlock()
		
		jsonBytes, err := json.Marshal(donor)
	
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write((jsonBytes))
		return;
	}
}

//updateUser 
func (h *usersHandler) updateUserContact(w http.ResponseWriter, r *http.Request,t string){
	println("You may only update contact Info")
	println("update user started \n if no_updation found, please check your field names: \n phone_no \n address");

	bodyBytes, err := ioutil.ReadAll(r.Body) //check body is valid
	defer r.Body.Close()

	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json"{
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("err:  content-type: application/json but got '%s'", ct)))
	}

	userId, err := strconv.Atoi(t);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid User id. Check Input UserId")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	userSecretCode, ok := h.store.IdsToSecretCodes[userId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input UserId")))
		return
	}
	
	h.Lock()
	userConfig, ok := h.store.SecretCodesToIds[userSecretCode];
	h.Unlock()

	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: UserId and Secret Code Mismatched. Contact Admin")))
		return;
	}

	var updateUser User;
	e := json.Unmarshal(bodyBytes, &updateUser) //updated user info
	if e != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return 
	}

	if userConfig.Type == Patient{
		h.Lock()
		currUser := h.store.Patients[userConfig.Id];
		if(updateUser.Address != ""){
			currUser.Address = updateUser.Address
		}
		if(updateUser.PhoneNo != ""){
			currUser.PhoneNo = updateUser.PhoneNo
		}
		h.store.Patients[userConfig.Id] = currUser;
		jsonBytes, err := json.Marshal(currUser)
	
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			h.Unlock()
			return
		}
		w.Header().Add("content-type", "application/json")
		w.Write((jsonBytes))
		h.Unlock();
	}

	if userConfig.Type == Donor{
		h.Lock()
		currUser := h.store.Donors[userConfig.Id];
		if(updateUser.Address != ""){
			currUser.Address = updateUser.Address
		}
		if(updateUser.PhoneNo != ""){
			currUser.PhoneNo = updateUser.PhoneNo
		}
		h.store.Donors[userConfig.Id] = currUser;
		jsonBytes, err := json.Marshal(currUser)
	
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			h.Unlock();
			return
		}
		w.Header().Add("content-type", "application/json")
		w.Write((jsonBytes))
		h.Unlock();
	}
	w.WriteHeader(http.StatusOK)
}

//deleteUser
func (h *usersHandler) deleteUser(w http.ResponseWriter, r *http.Request,t string){
	fmt.Println("\n delete started ");

	userId, err := strconv.Atoi(t);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid User id. Check Input UserId")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	userSecretCode, ok := h.store.IdsToSecretCodes[userId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input UserId")))
		return
	}
	
	h.Lock()
	userConfig, ok := h.store.SecretCodesToIds[userSecretCode];
	h.Unlock()
	
	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: UserId and Secret Code Mismatched. Contact Admin")))
		return;
	}

	if userConfig.Type == Patient{
		h.Lock()
		delete(h.store.Patients, userConfig.Id);
		delete(h.store.SecretCodesToIds, userSecretCode);
		delete(h.store.IdsToSecretCodes, userConfig.Id);
		h.store.Total -= 1;
		h.store.TotalPatients -= 1;
		h.Unlock();
	}

	if userConfig.Type == Donor{
		h.Lock()
		delete(h.store.Donors, userConfig.Id);
		delete(h.store.SecretCodesToIds, userSecretCode);
		delete(h.store.IdsToSecretCodes, userConfig.Id);
		h.store.Total -= 1;
		h.store.TotalDonors -= 1;
		h.Unlock();
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("user deleted. id: %d", userId)))
}


// sendRequest
func (h *usersHandler) sendRequest(w http.ResponseWriter, r *http.Request,t string , p string){
	fmt.Println("\n send Request started ");
	
	//user
	userId, err := strconv.Atoi(t);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid User id. Check Input UserId")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	userSecretCode, ok := h.store.IdsToSecretCodes[userId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input UserId")))
		return
	}
	
	h.Lock()
	userConfig, ok := h.store.SecretCodesToIds[userSecretCode];
	h.Unlock()

	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: Something is wrong! UserId and Secret Code Mismatched. Contact Admin")))
		return;
	}

	//request
	requestId, err := strconv.Atoi(p);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid Request id. Check request Id")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	requestSecretCode, ok := h.store.IdsToSecretCodes[requestId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input RequestId")))
		return
	}
	
	h.Lock()
	requestConfig, ok := h.store.SecretCodesToIds[requestSecretCode];
	h.Unlock()

	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: Something is wrong! RequestId and Secret Code Mismatched. Contact Admin")))
		return;
	}

	if userConfig.Type == Patient{
		h.Lock()
		currUser := h.store.Patients[userConfig.Id];
		isPresent := find(currUser.RequestedUserIds, requestId);
		
		if(isPresent!=-1){
			w.WriteHeader(http.StatusOK)
			h.Unlock()
			return
		}else{
			currUser.RequestedUserIds = append(currUser.RequestedUserIds, requestConfig.Id);
		}
		
		donorInfo, ok := h.store.Donors[requestConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId DonorId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}
		if find(donorInfo.PendingUserIds, userConfig.Id) == -1{
			donorInfo.PendingUserIds = append(donorInfo.PendingUserIds, userConfig.Id)
		}
		h.store.Donors[requestConfig.Id] = donorInfo;
		h.store.Patients[userConfig.Id] = currUser;
		h.Unlock();
	}

	if userConfig.Type == Donor{
		h.Lock()
		currUser := h.store.Donors[userConfig.Id];
		isPresent := find(currUser.RequestedUserIds, requestId);
		
		if(isPresent!=-1){
			w.WriteHeader(http.StatusOK)
			h.Unlock()
			return
		}else{
			currUser.RequestedUserIds = append(currUser.RequestedUserIds, requestConfig.Id);
		}
		
		patientInfo, ok := h.store.Patients[requestConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId PatientId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}
		if find(patientInfo.PendingUserIds, userConfig.Id) == -1{
			patientInfo.PendingUserIds = append(patientInfo.PendingUserIds, userConfig.Id)
		}
		
		h.store.Patients[requestConfig.Id] = patientInfo;
		h.store.Donors[userConfig.Id] = currUser;
		h.Unlock();
	}
	println("Requests Succesful")
	w.WriteHeader(http.StatusOK);
	return
}

//acceptRequest
func (h *usersHandler) acceptRequest(w http.ResponseWriter, r *http.Request,t string, p string ){
	fmt.Println("\n accept Request started ");
	//user
	userId, err := strconv.Atoi(t);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid User id. Check Input UserId")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	userSecretCode, ok := h.store.IdsToSecretCodes[userId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input UserId")))
		return
	}
	
	h.Lock()
	userConfig, ok := h.store.SecretCodesToIds[userSecretCode];
	h.Unlock()

	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: Something is wrong! UserId and Secret Code Mismatched. Contact Admin")))
		return;
	}

	//request
	requestId, err := strconv.Atoi(p);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid Request id. Check request Id")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	requestSecretCode, ok := h.store.IdsToSecretCodes[requestId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input RequestId")))
		return
	}
	
	h.Lock()
	requestConfig, ok := h.store.SecretCodesToIds[requestSecretCode];
	h.Unlock()

	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: Something is wrong! RequestId and Secret Code Mismatched. Contact Admin")))
		return;
	}

	if userConfig.Type == Patient{
		h.Lock()
		currUser := h.store.Patients[userConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. PatientId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		donorInfo, ok := h.store.Donors[requestConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. DonorId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		isPresent := find(currUser.ConnectedUsersIds, requestId);
		if(isPresent != -1){
			println("Connections Succesful")
			w.WriteHeader(http.StatusOK);
			return;
		}else{
			currUser.ConnectedUsersIds = append(currUser.ConnectedUsersIds, requestConfig.Id);
			donorInfo.ConnectedUsersIds = append(donorInfo.ConnectedUsersIds, userConfig.Id)
		}

		findRequestIndex := find(donorInfo.RequestedUserIds, userConfig.Id);
		if findRequestIndex == -1{
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("err: Unauthorized Connection Attempt. No Donor Request Sent by donorId: %d", requestConfig.Id)))
			h.Unlock();
			return
		}else{
			donorInfo.RequestedUserIds = removeElementByIndex(donorInfo.RequestedUserIds, findRequestIndex);
		}
		h.store.Donors[requestConfig.Id] = donorInfo;
		
		findRequestIndex = find(currUser.PendingUserIds, requestConfig.Id);
		if findRequestIndex == -1{
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("err: Unauthorized Connection Attempt. No Donor Request Received for donorId: %d", requestConfig.Id)))
			h.Unlock();
			return
		}else{
			currUser.PendingUserIds = removeElementByIndex(currUser.PendingUserIds, findRequestIndex);
		}
		
		h.store.Patients[userConfig.Id] = currUser;
		h.Unlock();
	}

	if userConfig.Type == Donor{
		h.Lock()
		currUser := h.store.Donors[userConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. PatientId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		patientInfo, ok := h.store.Patients[requestConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. PartentId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}
		
		isPresent := find(currUser.ConnectedUsersIds, requestId);
		if(isPresent != -1){
			println("Connections Succesful")
			w.WriteHeader(http.StatusOK);
			return;
		}else{
			currUser.ConnectedUsersIds = append(currUser.ConnectedUsersIds, requestConfig.Id);
			patientInfo.ConnectedUsersIds = append(patientInfo.ConnectedUsersIds, userConfig.Id)
		}

		findRequestIndex := find(patientInfo.RequestedUserIds, userConfig.Id);
		if findRequestIndex == -1{
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("err: Unauthorized Connection Attempt. No Donor Request Sent by donorId: %d", requestConfig.Id)))
			h.Unlock();
			return
		}else{
			patientInfo.RequestedUserIds = removeElementByIndex(patientInfo.RequestedUserIds, findRequestIndex);
		}
		h.store.Patients[requestConfig.Id] = patientInfo;
	
		findRequestIndex = find(currUser.PendingUserIds, requestConfig.Id);
		if findRequestIndex == -1{
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("err: Unauthorized Connection Attempt. No Donor Request Received Found for donorId: %d",requestConfig.Id)))
			h.Unlock();
			return
		}else{
			currUser.PendingUserIds = removeElementByIndex(currUser.PendingUserIds, findRequestIndex);
		}
	
		h.store.Donors[userConfig.Id] = currUser;
		h.Unlock();
	}
	println("Connections Succesful")
	w.WriteHeader(http.StatusOK);
	return
}

//cancelRequest
func (h *usersHandler) cancelRequest(w http.ResponseWriter, r *http.Request,t string, p string ){
	fmt.Println("\n request cancel started ");
	//user
	userId, err := strconv.Atoi(t);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid User id. Check Input UserId")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	userSecretCode, ok := h.store.IdsToSecretCodes[userId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input UserId")))
		return
	}
	
	h.Lock()
	userConfig, ok := h.store.SecretCodesToIds[userSecretCode];
	h.Unlock()

	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: Something is wrong! UserId and Secret Code Mismatched. Contact Admin")))
		return;
	}

	//request
	requestId, err := strconv.Atoi(p);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid Request id. Check request Id")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	requestSecretCode, ok := h.store.IdsToSecretCodes[requestId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input RequestId")))
		return
	}
	
	h.Lock()
	requestConfig, ok := h.store.SecretCodesToIds[requestSecretCode];
	h.Unlock()

	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: Something is wrong! RequestId and Secret Code Mismatched. Contact Admin")))
		return;
	}

	if(userConfig.Type == Patient){
		h.Lock()
		currUser := h.store.Patients[userConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. PatientId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		donorInfo, ok := h.store.Donors[requestConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. DonorId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		isPresent := find(currUser.RequestedUserIds, requestConfig.Id);
		if(isPresent == -1){
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("err: INVALID. No Send Request Found to DonorId : %d", requestConfig.Id)))
			h.Unlock();
			return
		}else{
			currUser.RequestedUserIds = removeElementByIndex(currUser.RequestedUserIds, isPresent);
		}
		h.store.Patients[userConfig.Id] = currUser;
		
		isPresent = find(donorInfo.PendingUserIds, userConfig.Id);
		donorInfo.PendingUserIds = removeElementByIndex(donorInfo.PendingUserIds, isPresent);
		
		h.store.Donors[requestConfig.Id] = donorInfo;
		h.Unlock();
	}
	if(userConfig.Type == Donor){
		h.Lock()
		currUser := h.store.Donors[userConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. UserId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		patientInfo, ok := h.store.Patients[requestConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. PatientId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		isPresent := find(currUser.RequestedUserIds, requestConfig.Id);
		if(isPresent == -1){
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("err: INVALID. No Send Request Found to PatientId : %d", requestConfig.Id)))
			h.Unlock();
			return
		}else{
			currUser.RequestedUserIds = removeElementByIndex(currUser.RequestedUserIds, isPresent);
		}
		h.store.Donors[userConfig.Id] = currUser;
		
		isPresent = find(patientInfo.PendingUserIds, userConfig.Id);
		patientInfo.PendingUserIds = removeElementByIndex(patientInfo.PendingUserIds, isPresent);
		
		h.store.Patients[requestConfig.Id] = patientInfo;
		h.Unlock();
	}
	println("Request Cancelled")
	w.WriteHeader(http.StatusOK);
	return
}

//cancelConnection
func (h *usersHandler) cancelConnection(w http.ResponseWriter, r *http.Request,t string, p string ){
	fmt.Println("\n connection cancel started ");
	//user
	userId, err := strconv.Atoi(t);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid User id. Check Input UserId")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	userSecretCode, ok := h.store.IdsToSecretCodes[userId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input UserId")))
		return
	}
	
	h.Lock()
	userConfig, ok := h.store.SecretCodesToIds[userSecretCode];
	h.Unlock()

	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: Something is wrong! UserId and Secret Code Mismatched. Contact Admin")))
		return;
	}

	//request
	requestId, err := strconv.Atoi(p);
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("err: invalid Request id. Check request Id")))
		return
	}

	fmt.Println(userId);
	
	h.Lock();
	requestSecretCode, ok := h.store.IdsToSecretCodes[requestId];
	h.Unlock();
	
	if !ok{
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("err: User Not Found. Check Input RequestId")))
		return
	}
	
	h.Lock()
	requestConfig, ok := h.store.SecretCodesToIds[requestSecretCode];
	h.Unlock()

	if !ok{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("err: Something is wrong! RequestId and Secret Code Mismatched. Contact Admin")))
		return;
	}

	if(userConfig.Type == Patient){
		h.Lock()
		currUser := h.store.Patients[userConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. PatientId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		donorInfo, ok := h.store.Donors[requestConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. DonorId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		isPresent := find(currUser.ConnectedUsersIds, requestConfig.Id);
		if(isPresent == -1){
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("err: INVALID. No Connection Found b/w DonorId : %d and UserID: %d", requestConfig.Id, userConfig.Id)))
			h.Unlock();
			return
		}else{
			currUser.ConnectedUsersIds = removeElementByIndex(currUser.ConnectedUsersIds, isPresent);
		}
		h.store.Patients[userConfig.Id] = currUser;
		
		isPresent = find(donorInfo.ConnectedUsersIds, userConfig.Id);
		donorInfo.ConnectedUsersIds = removeElementByIndex(donorInfo.ConnectedUsersIds, isPresent);
		h.store.Donors[requestConfig.Id] = donorInfo;
		h.Unlock();
	}
	if(userConfig.Type == Donor){
		h.Lock()
		currUser := h.store.Donors[userConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. UserId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		patientInfo, ok := h.store.Patients[requestConfig.Id];
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("err: Check Input UserId. PatientId : %d NOT FOUND", requestConfig.Id)))
			h.Unlock();
			return
		}

		isPresent := find(currUser.ConnectedUsersIds, requestConfig.Id);
		if(isPresent == -1){
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("err: INVALID. No Send Request Found to PatientId : %d", requestConfig.Id)))
			h.Unlock();
			return
		}else{
			currUser.ConnectedUsersIds = removeElementByIndex(currUser.ConnectedUsersIds, isPresent);
		}
		h.store.Donors[userConfig.Id] = currUser;
		
		isPresent = find(patientInfo.ConnectedUsersIds, userConfig.Id);
		patientInfo.ConnectedUsersIds = removeElementByIndex(patientInfo.ConnectedUsersIds, isPresent);

		h.store.Patients[requestConfig.Id] = patientInfo;
		h.Unlock();
	}
	println("Connection Cancelled")
	w.WriteHeader(http.StatusOK);
	return
}

//func init
func main(){
	usersHandler := newUsersHandler();
	http.HandleFunc("/users/", usersHandler.users);
	http.HandleFunc("/user/",usersHandler.user);

	err := http.ListenAndServe(":8080", nil);
	if err != nil{
		panic(err)
	}
}