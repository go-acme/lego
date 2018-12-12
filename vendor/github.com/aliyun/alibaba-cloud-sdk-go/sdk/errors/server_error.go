/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package errors

import "fmt"

type ServerError struct {
	httpStatus int
	errorCode  string
	message    string
}

func NewServerError(httpStatus int, errorCode, message string) Error {
	return &ServerError{
		httpStatus: httpStatus,
		errorCode:  errorCode,
		message:    message,
	}
}

func (err *ServerError) HttpStatus() int {
	return err.httpStatus
}

func (err *ServerError) ErrorCode() string {
	return err.errorCode
}

func (err *ServerError) Message() string {
	return err.message
}

func (err *ServerError) Error() string {
	return fmt.Sprintf("SDK.ServerError %s %s", err.errorCode, err.message)
}

func (err *ServerError) OriginError() error {
	return nil
}
