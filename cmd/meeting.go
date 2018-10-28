// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	. "github.com/wtysos11/goAgenda/entity"
	"github.com/spf13/cobra"
	"errors"
)

const meetingPlace = "meeting.txt"
//check whether all the users are available and have time to attend this meeting
func userTimeCheck(userInfo []User,meetingInfo []Meeting,startTime AgendaTime, endTime AgendaTime,participants []string) error{
	//first, check all participants are available in userInfo
	for _,p := range participants {
		pass := false
		for _,u := range userInfo{
			if u.Username == p{
				pass = true
				break
			}
		}
		if(!pass){
			return errors.New("Participants have illegal participant:"+p)
		}
	}
	//for all meetings, if their userlist have participant, check whether this meeting have conflicts.
	for _,m := range meetingInfo{
		inMeeting := false
		var involveParticipant string
		for _,user := range m.UserList{
			for _,p := range participants {
				if user == p{
					inMeeting = true
					involveParticipant = p
					break
				}
			}
		}
		if !inMeeting{
			continue
		}

		meetingStartTime,_ := String2Time(m.StartTime)
		meetingEndTime,_ := String2Time(m.EndTime)
		if !(CompareTime(endTime,meetingStartTime)<0 && CompareTime(startTime,meetingEndTime)>0){
			return errors.New("For participants "+involveParticipant+". Meeting "+m.Title+" have conflicts.") 
		}
	}
	return nil
}

//legal check, don't implement yet
func meetingLegalCheck(meetingInfo []Meeting,startTime string, endTime string,title string ,participants []string) (bool,error){
	sTime,tserr := String2Time(startTime)
	eTime,teerr := String2Time(endTime)
	if tserr!=nil {
		return false,tserr
	} else if teerr != nil{
		return false,teerr
	}

	if startTimeErr := TimeLegalCheck(sTime); startTimeErr != nil{
		return false,startTimeErr
	} 
	if endTimeErr := TimeLegalCheck(eTime); endTimeErr != nil{
		return false,endTimeErr
	}

	if CompareTime(sTime,eTime)>=0 {
		return false,errors.New("start time should smaller than end time (equal is not allowed)")
	}

	for _,m := range meetingInfo {
		if m.Title == title{
			return false,errors.New("Repeat title")
		}
	}

	//check participants
	if userInfo,userReadingerr := ReadUserFromFile(userPlace);userReadingerr!=nil {
		fmt.Println(userReadingerr)
		return false,userReadingerr
	} else{
		if userCheckError := userTimeCheck(userInfo,meetingInfo,sTime,eTime,participants);userCheckError != nil{
			return false,userCheckError
		}
	}

	return true,nil
}



// meetingCmd represents the meeting command
var meetingCmd = &cobra.Command{
	Use:   "meeting",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//all meeting operation need a login status
		if login,err:=checklogin(); err!=nil{
			fmt.Println(err)
			return
		} else if !login{
			fmt.Println("Please login first")
			return
		}
		//get login username, for founder of the conference and some operations
		loginUsername,loginErr := getLoginUsername()
		if loginErr !=nil {
			fmt.Println(loginErr)
			return
		}

		//get store meeting data in JSON format
		meetingInfo,meetingReadingerr := ReadMeetingFromFile(meetingPlace)
		if meetingReadingerr!=nil {
			fmt.Println(meetingReadingerr)
			return
		}
		fmt.Println("meeting called")
		startTime,_ := cmd.Flags().GetString("start")
		endTime,_ := cmd.Flags().GetString("end")
		title,_ := cmd.Flags().GetString("title")
		participants,_ := cmd.Flags().GetStringArray("participant")

		/*
		fmt.Println("flags test.")
		fmt.Println(startTime)
		fmt.Println(endTime)
		fmt.Println(title)
		fmt.Println(participants)
*/
		if len(args)>0{
			switch (args[0]){
				case "create":{
					fmt.Println("create")

					if pass,err := meetingLegalCheck(meetingInfo,startTime,endTime,title,participants); err !=nil {
						fmt.Println(err)
						return
					} else if !pass{
						fmt.Println("Meeting create failed")
						return
					}
					meetingInfo = append(meetingInfo,Meeting{loginUsername,startTime,endTime,title,participants})

					WriteMeetingToFile(meetingPlace,meetingInfo)
					fmt.Println("Meeting create success")
				}
				case "addUser":{
					fmt.Println("add user")
					//check. Need title and at least one valiable participants(username correct and have time to attend)

					//find meeting
					pass := false
					for i , meeting := range meetingInfo{
						if meeting.Title == title{
							pass = true
							//check whether participants have time

							meetingInfo[i].UserList = append(meetingInfo[i].UserList,participants...)
							break
						}
					}

					if !pass {
						fmt.Println("Meeting add users failed.")
						return
					}
					WriteMeetingToFile(meetingPlace,meetingInfo)
					fmt.Println("Meeting add users success")
				}
				case "deleteUser":{
					fmt.Println("delete user")
					//check. title and participants name

					//find meeting
					pass := false
					for i := 0; i < len(meetingInfo);i++ {
						meeting := meetingInfo[i]
						if meeting.Title == title{ //find the meeting
							pass = true
							//check whether participants have time

							//delete participants from this meeting
							//warning: may have bugs. Not sure
							for j := 0; j < len(meeting.UserList) ; j++ {
								user := meeting.UserList[j]
								for k:=0 ; k < len(participants) ; k++ {
									deleteUser:=participants[k]
									if user == deleteUser{
										if j+1 < len(meetingInfo[i].UserList) {
											meetingInfo[i].UserList = append(meetingInfo[i].UserList[:j],meetingInfo[i].UserList[j+1:]...)
											j--;
										} else {
											meetingInfo[i].UserList = meetingInfo[i].UserList[:j]
										}
										
										if k+1 < len(participants) {
											participants = append(participants[:k],participants[k+1:]...)
											k--;
										} else {
											participants = participants[:k]
										}
										
										break
									}
								}
							}
							//if the delete operation make this meeting empty, clear the meeting
							if len(meetingInfo[i].UserList) == 0 {
								if i+1 < len(meetingInfo){
									meetingInfo = append(meetingInfo[:i],meetingInfo[i+1:]...)
									i--;
								} else {
									meetingInfo = meetingInfo[:i]
								}
								
							}
							break
						}
					}

					if !pass {
						fmt.Println("Meeting delete users failed.")
						return
					}
					WriteMeetingToFile(meetingPlace,meetingInfo)
					fmt.Println("Meeting delete users success")
				}
				case "lookup":{
					fmt.Println("meeting lookup")
				}
				case "cancel":{
					fmt.Println("meeting cancel")

					pass := false
					for i := 0 ; i<len(meetingInfo) ; i++{
						meeting := meetingInfo[i]
						if meeting.Title == title && meeting.Creator == loginUsername{
							pass = true
							if(i+1<len(meetingInfo)){
								meetingInfo = append(meetingInfo[:i],meetingInfo[i+1:]...)
								i--
							} else{
								meetingInfo = meetingInfo[:i]
							}
							break
							
						} else if meeting.Title == title && meeting.Creator != loginUsername{
							fmt.Println("You can only cancel the meeting that create by yourself")
							return
						}
					}
					if !pass{
						fmt.Println("Meeting cancel failed")
						return
					}
					WriteMeetingToFile(meetingPlace,meetingInfo)
					fmt.Println("Meeting cancel success")
				}
				case "exit":{
					fmt.Println("meeting exit")
					//find the specific meeting

					//check whether user join
					pass := false
					for i:=0 ; i < len(meetingInfo) ; i++ {
						meeting := meetingInfo[i]
						if meeting.Title == title{
							//check whether the user in the user list
							for j := 0 ; j < len(meeting.UserList) ; j++ {
								user := meeting.UserList[j]
								if user == loginUsername{
									pass = true
									if(j+1 < len(meetingInfo[i].UserList)){
										meetingInfo[i].UserList = append(meetingInfo[i].UserList[:j],meetingInfo[i].UserList[j+1:]...)
										j--
									} else{
										meetingInfo[i].UserList = meetingInfo[i].UserList[:j]
									}
									// if UserList is empty, delete the meeting
									if len(meetingInfo[i].UserList) == 0{
										if i+1 < len(meetingInfo){
											meetingInfo = append(meetingInfo[:i],meetingInfo[i+1:]...)
										} else{
											meetingInfo = meetingInfo[:i]
										}
									}

									break
								}
							}

							break
						}
					}

					if !pass {
						fmt.Println("Meeting Exit failed")
						return
					}

					//delete it
					WriteMeetingToFile(meetingPlace,meetingInfo)
					fmt.Println("Meeting Exit success")
				}
				case "clear":{
					fmt.Println("meeting clear")
				}
			}
		} else{
			fmt.Println("Need some command.")
		}
	},
}

func init() {
	rootCmd.AddCommand(meetingCmd)
	meetingCmd.Flags().StringP("start","s","2018-10-28/16:01:23","Help message for start time")
	meetingCmd.Flags().StringP("end","e","9999-12-31/23:59:59","Help message for end time")
	meetingCmd.Flags().StringP("title","t","conference","Help message for meeting title")
	meetingCmd.Flags().StringArrayP("participant","p",[]string{},"Help message for participant")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// meetingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// meetingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
